import { mount } from 'svelte'
import App from './App.svelte'
import './global.css'

export default mount(App, { target: document.body })
