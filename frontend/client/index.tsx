import { createRoot } from 'react-dom/client';
import {App} from './app';
import '../styles/output.css'

// Render the React component into the templ page at the react-header.
const root = document.getElementById('root');
if (!root) {
    throw new Error('Could not find element with id react-header');
}
const reactRoot = createRoot(root);
reactRoot.render(App());
