import {Link} from "react-router";

export const Header = () => {
    return (
        <div className={'bg-sky-500'}>
            <Link to={'/'}>Home</Link>
            <Link to={'/about'}>About</Link>
        </div>
    )
}
