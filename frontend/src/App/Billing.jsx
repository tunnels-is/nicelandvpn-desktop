import { useNavigate, Navigate } from "react-router-dom";
import React, { useEffect, useState } from "react";
import STORE from "../store";

const useForm = (props) => {

    const [inputs, setInputs] = useState({});
    const [user, setUser] = useState();

    const Submit = () => {
        console.log("SENDING INPUT")
        console.dir(input)
    }

    const UpdateInput = (index, key, value) => {

        let i = { ...inputs }
        let hasErrors = false

        i[index][key] = value

        if (value === "") {
            i[index].Errors[key] = key + " missing"
        } else {
            i[index].Errors[key] = ""
        }


        console.log("updating inputs")
        console.dir(i)
        setInputs(i)

        if (hasErrors) {
            props.toggleError("Error in settings input")
        }

    }

    return {
        inputs,
        setInputs,
        user,
        setUser,
        UpdateInput,
        Submit
    }
}

const Settings = (props) => {

    const navigate = useNavigate();
    const { inputs, setInputs, user, setUser, UpdateInput, submit } = useForm(props);
    const [invoices, setInvoices] = useState([]);

    const UpdatePasswordRequest = () => {
        console.log("SENDING PASSWORD UPDATE REQUEST")
    }

    useEffect(() => {
        const u = STORE.Cache.GetObject("user");

        if (!u) {
            return (<Navigate redirect to="/login" />)
        }
        setUser(u)

        let i = {}
        i["Email"] = {}
        i["Email"].Errors = {}
        i["Email"].Email = u.Email

        setInputs(i)


        // TODO
        // .. get 5 latest invoices (and cache for X days)
        let invs = [
            { _id: "1", Title: "1 Month Subscription", Amount: 5.99, Created: "19. January 1999" },
            { _id: "2", Title: "1 Month Subscription", Amount: 5.99, Created: "19. January 1999" },
            { _id: "3", Title: "1 Month Subscription", Amount: 5.99, Created: "19. January 1999" }
        ]

        STORE.Cache.SetObject("invoices", invs)
        setInvoices(invs)
    }, []);

    if (!user) {
        return (<></>)
    }
    return (
        <div className="settings-wrapper row">


            <div className="card billing-card">
                <h5 className="card-header">Billing</h5>
                <div className="card-body">
                    {invoices.map(invoice => (
                        <>
                            <b className="dimmed">{invoice.Title} <a className="pdf-link" href={"/invoice/" + invoice._id}>PDF</a></b>
                            <div className="card-value">{invoice.Amount} USD</div>
                            <div className="card-value">{invoice.Created}</div>
                        </>
                    ))}
                </div>
                <button disabled onClick={() => console.log("see all invoices")} className={`btn invoice-button`} > All Invoices </button>
            </div>


        </div >
    )
}

export default Settings;