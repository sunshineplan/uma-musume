import { Dexie } from 'dexie'
import { getCookie, setCookie, removeCookie } from 'typescript-cookie'

const db = new Dexie('umamusume')
db.version(1).stores({
  events: '++id',
  images: 'id',
})
const eventsTable = db.table<Event>('events')
const imagesTable = db.table<Image>('images')

const init = async (last: string) => {
  const resp = await fetch('uma.json', { cache: 'no-cache' })
  const events = await resp.json()
  if (events) {
    await db.transaction('rw!', eventsTable, async () => {
      await eventsTable.clear()
      await eventsTable.bulkAdd(events)
    }).then(() => setCookie('last', last, { expires: 365 }))
  }
  return events
}

const loadEvents = async () => {
  const resp = await fetch('last', { cache: 'no-cache' })
  const last = await resp.text()
  let events: Event[]
  if (last != getCookie('last')) events = await init(last)
  else events = await eventsTable.toArray()
  if (!events || !events.length) {
    removeCookie('last')
    return await loadEvents()
  }
  return events
}
const allEevents = await loadEvents()

export const characters: (FilterTypeRegistry['character'])[] =
  Array.from(allEevents.filter(event => event.t == 'c'), i => { return { name: i.c, image: i.i } })
    .concat(Array.from(allEevents.filter(event => event.t == 'm'), i => { return { name: i.c, image: i.i } }))
    .filter((obj, index, arr) => arr.findIndex(i => (i.name == obj.name)) == index)

class UMA {
  filter = $state<Filter<FilterType>>({ type: 'character', name: '', image: '' })
  support = $state<Support>({ rare: 'SSR' })
  query = $state('')
  count = $state(0)
  events = $derived.by(() => {
    let events = <Event[]>[]
    if (this.filter.name) {
      if (this.filter.type == 'character')
        events = allEevents.filter(event => (event.t == 'c' || event.t == 'm') && event.c == this.filter.name)
      else if (this.filter.type == 'support')
        events = allEevents.filter(event => event.t == 's' && event.i == this.filter.image)
    }
    else events = allEevents
    this.count = events.length
    events.sort((a, b) => {
      if (a.c == b.c)
        if (a.t == b.t)
          if (a.r == b.r)
            return a.e.localeCompare(b.e)
          else return a.r.localeCompare(b.r)
        else return a.t.localeCompare(b.t)
      else return a.c.localeCompare(b.c)
    })
    if (this.query) {
      const res = <Event[]>[]
      events.forEach((event) => {
        if (this.#match(uma.query, event)) res.push(event);
      });
      return res
    }
    else return events
  })
  supports = $derived.by<(FilterTypeRegistry['support'])[]>(() => {
    let supports = Array.from(allEevents.filter(event => event.t == 's'), i => { return { name: i.c, rare: i.r, image: i.i } })
      .filter((obj, index, arr) => arr.findIndex(i => (i.image == obj.image)) == index)
    if (this.support.type)
      supports = supports.filter(support => support.rare.includes(this.support.type!))
    if (this.support.rare)
      supports = supports.filter(support => support.rare.replace(/[^SR]+/g, '') == this.support.rare)
    supports.sort((a, b) => {
      if (a.rare.replace(/[^SR]+/g, '') == b.rare.replace(/[^SR]+/g, ''))
        return a.name.localeCompare(b.name)
      else return b.rare.replace(/[^SR]+/g, '').localeCompare(a.rare.replace(/[^SR]+/g, ''))
    })
    return supports
  })
  #match(value: string, event: Event) {
    if (
      event.c.includes(value) ||
      event.e.includes(value) ||
      event.k.includes(value)
    )
      return true
    let matched = false
    event.o.forEach((option) => {
      if (option.b.includes(value)) matched = true
    })
    return matched
  }
  async loadImage(id: string) {
    const res = await db.transaction("r", imagesTable, () => {
      return imagesTable.get({ id });
    });
    return res?.image
  }
  async saveImage(image: Image) {
    await imagesTable.put(image)
  }
}
export const uma = new UMA

class Toggler {
  status = $state(false)
  toggle() { this.status = !this.status }
  close() { this.status = false }
}
export const showFilter = new Toggler
