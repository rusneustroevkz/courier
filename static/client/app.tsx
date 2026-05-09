import {BrowserRouter, Route, Routes} from "react-router";
import Home from "./pages/home";
import Orders from "./pages/orders";

export const App = () => (
    <BrowserRouter>
        <Routes>
            <Route path="/" element={<Home />} />
            <Route path={"/about"} element={<Orders />} />
        </Routes>
    </BrowserRouter>
);