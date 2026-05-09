import {BrowserRouter, Link, Route, Routes} from "react-router";
import Home from "./pages/home";
import Orders from "./pages/orders";
import {SidebarProvider, SidebarTrigger} from "@/components/ui/sidebar";
import {AppSidebar} from "./app-sidebat";

export const App = () => (
    <BrowserRouter>
        <SidebarProvider>
            <AppSidebar />
            <main>
                <SidebarTrigger />
                <Routes>
                    <Route path="/" element={<Home />} />
                    <Route path={"/about"} element={<Orders />} />
                </Routes>
            </main>
        </SidebarProvider>
    </BrowserRouter>
);