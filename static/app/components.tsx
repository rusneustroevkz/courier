export const App = () => (
    <div>
        <h1>React component Header</h1>
        <Body></Body>
        <Hello name={"asd"}></Hello>
    </div>
);

export const Body = () => (<div>This is client-side content from React</div>);

export const Hello = ({name: string}) => (
    <div>Hello {name} (Client-side React, rendering server-side data)</div>
);