interface Event {
  e: string      // Event
  c: string      // Character
  t: string      // Type
  r: string      // Rare
  a: string      // Article
  i: string      // Image
  k: string      // Keyword
  o: {
    b: string    // Branch
    g: string    // Gain
    s: object    // Skill
  }[]            // Options
}

interface Support {
  type?: 'スピ' | 'スタ' | 'パワ' | '根性' | '賢さ' | '友人' | 'グル'
  rare?: 'SSR' | 'SR' | 'R'
}
