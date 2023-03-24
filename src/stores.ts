import type { Writable, Readable } from 'svelte/store'
import { writable, derived } from 'svelte/store'
import { Dexie } from 'dexie'
import { getCookie, setCookie, removeCookie } from 'typescript-cookie'

interface FilterTypeRegistry {
  character: {
    name: string
    rare?: string
    image: string
  }
  support: {
    name: string
    rare: string
    image: string
  }
}

type FilterType = keyof FilterTypeRegistry

type Filter<FType extends FilterType = FilterType> = { type: FType } & FilterTypeRegistry[FType]

export const db = new Dexie('umamusume')
db.version(1).stores({
  events: '++id',
  images: 'id',
})

const init = async (last: string) => {
  await db.table('events').clear()
  const events = await fetch('uma.json', { cache: 'no-cache' })
  await db.table('events').bulkAdd(await events.json())
  setCookie('last', last, { expires: 365 })
}

const loadEvents = async (): Promise<Event[]> => {
  const resp = await fetch('last', { cache: 'no-cache' })
  const last = await resp.text()
  if (last != getCookie('last')) await init(last)
  const events = await db.table('events').toArray()
  if (!events.length) {
    removeCookie('last')
    return await loadEvents()
  }
  return events
}
const uma = await loadEvents()

export const characters: (FilterTypeRegistry['character'] & { image: string })[] =
  Array.from(uma.filter(event => event.t == 'c'), i => { return { name: i.c, image: i.i } })
    .concat(Array.from(uma.filter(event => event.t == 'm'), i => { return { name: i.c, image: i.i } }))
    .filter((obj, index, arr) => arr.findIndex(i => (i.name == obj.name)) == index)

export const filter: Writable<Filter<FilterType>> = writable({ type: 'character', name: '', image: '' })
export const support: Writable<Support> = writable({ rare: 'SSR' })
export const query = writable('')

export const events: Readable<Event[]> = derived(filter, $filter => {
  let events: Event[] = []
  if ($filter.name) {
    if ($filter.type == 'character')
      events = uma.filter(event => (event.t == 'c' || event.t == 'm') && event.c == $filter.name)
    else if ($filter.type == 'support')
      events = uma.filter(event => event.t == 's' && event.i == $filter.image)
  }
  else events = uma
  events.sort((a, b) => {
    if (a.c == b.c)
      if (a.t == b.t)
        if (a.r == b.r)
          return a.e.localeCompare(b.e)
        else return a.r.localeCompare(b.r)
      else return a.t.localeCompare(b.t)
    else return a.c.localeCompare(b.c)
  })
  return events
})

export const supports: Readable<(FilterTypeRegistry['support'] & { image: string })[]> = derived(support, $support => {
  let supports = Array.from(uma.filter(event => event.t == 's'), i => { return { name: i.c, rare: i.r, image: i.i } })
    .filter((obj, index, arr) => arr.findIndex(i => (i.image == obj.image)) == index)
  if ($support.type)
    supports = supports.filter(support => support.rare.includes($support.type as string))
  if ($support.rare)
    supports = supports.filter(support => support.rare.replace(/[^SR]+/g, '') == $support.rare)
  supports.sort((a, b) => {
    if (a.rare.replace(/[^SR]+/g, '') == b.rare.replace(/[^SR]+/g, ''))
      return a.name.localeCompare(b.name)
    else return b.rare.replace(/[^SR]+/g, '').localeCompare(a.rare.replace(/[^SR]+/g, ''))
  })
  return supports
})

const createShowFilter = () => {
  const { subscribe, set, update } = writable(false)
  return {
    subscribe,
    switch: () => update(status => !status),
    off: () => set(false)
  }
}
export const showFilter = createShowFilter()
