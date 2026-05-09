import {Header} from "./ui/header";
import {BrowserRouter, Route, Routes} from "react-router";
import Home from "./pages/home";
import About from "./pages/about";

export const App = () => (
    <BrowserRouter>
        <Header />
        <Routes>
            <Route path="/" element={<Home />} />
            <Route path={"/about"} element={<About />} />
        </Routes>
    </BrowserRouter>
);