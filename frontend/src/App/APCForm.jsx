import React, { useEffect, useState } from "react";
import STORE from "../store";

const APCREATE = (props) => {

    const customState = (props) => {
        const [inputs, setInputs] = useState({ Networks: [], Access: [], Tag: "", Region: "" });
        const [dynamicCounter, setDynamicCounter] = useState({ Count: 0 });
        const [changed, setChanged] = useState(false)


        const handleDelete = async (ID) => {
            console.log("REMOVING AP WITH ID", ID)
        }

        const handleSubmit = async (event) => {
            if (event) {
                event.preventDefault();
            }

            if (!changed) {
                props.toggleError("You have not made any changed to the form")
                return
            }

            console.log("SUBMITTING")
            console.dir(inputs)
            console.log("SUBMITTING")

            let user = STORE.Cache.GetObject("user")
            let form = {}
            let hasError = false

            form.DeviceToken = user.DeviceToken.DT

            form.AP = {}
            form.AP.UID = user._id
            form.AP.Tag = inputs.Tag
            form.AP.InternetAccess = inputs.InternetAccess
            form.AP.Access = []
            form.AP.Networks = []
            form.AP.Region = inputs.Region

            inputs["Access"].map(acc => {
                if (!acc.UID || acc.UID === "") {
                    hasError = true
                }
                if (!acc.Tag || acc.Tag === "") {
                    hasError = true
                }
                form.AP.Access.push({ UID: acc.UID, Tag: acc.Tag })
            })

            inputs["Networks"].map(net => {
                if (!net.Network || net.Network === "") {
                    hasError = true
                }
                if (!net.LocalNetwork || net.LocalNetwork === "") {
                    hasError = true
                }
                if (!net.Tag || net.Tag === "") {
                    hasError = true
                }
                form.AP.Networks.push({ Tag: net.Tag, Network: net.Network, LocalNetwork: net.LocalNetwork })
            })


            if (hasError) {
                props.toggleError("VPN can not be saved with errors")
                return
            }

            let FR = {
                Path: "user/device/create",
                Method: "POST",
                JSONData: form,
                Timeout: 20000
            }

            console.dir(form)

            try {
                // const response = await CLIENT.post("forward", JSON.stringify(FR));
                // console.dir(response)

                // setAP({ ...a })
                // let ad = STORE.Cache.GetObject("advanced-data-private")
                // ad.APS.map((ap) => {
                //     if (ap.ConnectKey === AP.ConnectKey) {
                //         ap.Tag = form.Tag
                //         ap.Access = form.Access
                //         ap.Networks = form.Networks
                //         ap.InternetAccess = form.InternetAccess
                //         ap.Region = form.RG
                //     }
                // })
                // STORE.Cache.SetObject("advanced-data-private", ad)

                props.showSuccessToast("VPN created.. it will take a minute or two to show up in your VPN list")
            } catch (error) {
                props.toggleError(STORE.ParseResponseErrorMessage(error))
            }


        }

        const UpdateInput = (type, index, key, value) => {

            let i = { ...inputs }
            let hasErrors = false


            if (type === "Access") {
                i.Access[index][key] = value
                if (value === "") {
                    if (key === "UID") {
                        i.Access[index].Errors[key] = "User ID missing"
                    } else {
                        i.Access[index].Errors[key] = key + " missing"
                    }
                } else {
                    delete i.Access[index].Errors[key]
                }

            } else if (type === "Networks") {
                i.Networks[index][key] = value
                if (value === "") {
                    if (key === "Network") {
                        i.Networks[index].Errors[key] = key + " missing"
                    } else if (key === "LocalNetwork") {
                        i.Networks[index].Errors[key] = key + " missing"
                    } else if (key === "Tag") {
                        i.Networks[index].Errors[key] = "Tag missing"
                    }
                } else {
                    delete i.Networks[index].Errors[key]
                }

            } else if (type === "Tag") {
                i.Tag = value
            } else if (type === "InternetAccess") {
                i.InternetAccess = value
            } else if (type === "Region") {
                i.Region = value
            } else if (type === "LocalNetworkAccess") {
                i.LocalNetworkAccess = value
            }


            setChanged(true)
            setInputs(i)

            if (hasErrors) {
                props.toggleError("VPN inputs has errors")
            }

        }

        const RemoveField = (type, index) => {
            // console.log("REMOVING FIELD:", index)
            let i = { ...inputs }
            if (type === "Networks") {
                delete i.Networks[index]
            } else if (type === "Access") {
                delete i.Access[index]
            }

            setInputs(i)
        }

        return {
            inputs,
            setInputs,
            dynamicCounter,
            setDynamicCounter,
            UpdateInput,
            handleSubmit,
            handleDelete,

            RemoveField
        };
    }

    const { inputs, setInputs, dynamicCounter, setDynamicCounter, UpdateInput, handleSubmit, handleDelete, RemoveField } = customState(props);


    // let { id } = useParams()

    // let GEOKeys = []
    // let GEOExtraKeys = []
    // let BasicKeys = []
    // let StatsKeys = []
    // let IDs = []
    // let AccessInputFields = []
    // let NetworkInputFields = []

    useEffect(() => {
        // let inputs = {}
        // inputs.Access = []
        // inputs.Networks = []
        // let data = STORE.Cache.GetObject("advanced-data-private")
        // let ldc = { ...dynamicCounter }
        // console.log(id)
        // accesspoints = data.APS
        // if (accesspoints !== undefined) {
        //     accesspoints.map(ap => {
        //         console.log("comparing APS", ap._id, id)
        //         if (ap.ConnectKey === id) {
        //             // if (ap._id === id) {
        //             setAP(ap)
        //             inputs.Tag = ap.Tag
        //             inputs.InternetAccess = ap.InternetAccess
        //             inputs.RG = ap.Region

        //             if (ap.Access) {
        //                 ap.Access.map(acc => {
        //                     ldc.Count = ldc.Count + 1
        //                     inputs.Access[ldc.Count] = {}
        //                     inputs.Access[ldc.Count].UID = acc.UID
        //                     inputs.Access[ldc.Count].Tag = acc.Tag
        //                     inputs.Access[ldc.Count].Errors = {}
        //                 })
        //             }
        //             if (ap.Networks) {
        //                 ap.Networks.map(n => {
        //                     ldc.Count = ldc.Count + 1
        //                     inputs.Networks[ldc.Count] = {}
        //                     inputs.Networks[ldc.Count].Network = n.Network
        //                     inputs.Networks[ldc.Count].LocalNetwork = n.LocalNetwork
        //                     inputs.Networks[ldc.Count].Tag = n.Tag
        //                     inputs.Networks[ldc.Count].Errors = {}
        //                 })
        //             }
        //         }
        //     })

        // setInputs(inputs)
        //     setDynamicCounter(ldc)
        // }
    }, [])

    const AddAccessField = () => {
        let i = { ...inputs }
        let ldc = { ...dynamicCounter }
        ldc.Count = ldc.Count + 1
        i.Access[ldc.Count] = {}
        i.Access[ldc.Count].UID = ""
        i.Access[ldc.Count].Tag = ""
        i.Access[ldc.Count].Errors = { "Tag": "Tag missing", "UID": "User ID missing" }

        setDynamicCounter(ldc)
        setInputs(i)
    }
    const AddNetworkField = () => {
        let i = { ...inputs }
        let ldc = { ...dynamicCounter }
        ldc.Count = ldc.Count + 1
        i.Networks[ldc.Count] = {}
        i.Networks[ldc.Count].Network = "<Network Mapping>"
        i.Networks[ldc.Count].LocalNetwork = "<Local Network>"
        i.Networks[ldc.Count].Tag = "<Tag>"
        i.Networks[ldc.Count].Errors = {}

        setDynamicCounter(ldc)
        setInputs(i)
    }




    // console.dir(errors)
    console.log("global inputs")
    console.dir(inputs)
    // console.dir(AP)
    // console.dir(GEOExtraKeys)
    // console.dir(GEOKeys)
    return (
        <div className="ap-form-wrapper">
            <div className="info panel">
                <div className="title">Info</div>
                <div className="save" onClick={handleSubmit} >Save</div>
                <div className="seperator"></div>


                <div className="row">
                    <div className="label">Tag</div>
                    <div className="edit"></div>
                    <input className="input editable" onChange={e => UpdateInput("Tag", 0, "Tag", e.target.value)} value={inputs["Tag"]} />
                </div>
                <div className="row">
                    <div className="label">Region</div>
                    <div className="edit"></div>

                    <select className="input select editable" onChange={e => UpdateInput("Region", 0, "Region", e.target.value)}>
                        <option selected disabled className="option">Select Region</option>
                        {STORE.Config.REGIONS.map(r => {
                            return (
                                <option className="option" value={r.value}>{r.desc}</option>
                            )
                        })}
                    </select>
                    {/* <input className="input editable" onChange={e => UpdateInput("Region", 0, "Region", e.target.value)} value={inputs["Region"]} /> */}
                </div>



            </div>

            <div className="network panel">
                <div className="title">NAT / Network Mapping</div>
                <div className="add" onClick={() => AddNetworkField()} >+</div>
                <div className="seperator"></div>

                {inputs["Networks"].map((net, index) => {
                    return (
                        <>
                            <div className={`network-card-${index}`} key={index}>


                                {net.Errors["Tag"] && <b className="error">{net.Errors["Tag"]}</b>}

                                <div className="row">
                                    <div className="label">Tag</div>
                                    <div className="edit"></div>
                                    <input className="input editable" onChange={e => UpdateInput("Networks", index, "Tag", e.target.value)} value={net.Tag} />
                                </div>

                                {net.Errors["Network"] && <b className="error">{net.Errors["Network"]}</b>}

                                <div className="row" >
                                    <div className="label">Mapping</div>
                                    <div className="edit"></div>
                                    <input className="input editable" onChange={e => UpdateInput("Networks", index, "Network", e.target.value)} value={net.Network} />
                                </div>

                                {net.Errors["LocalNetwork"] && <b className="error">{net.Errors["LocalNetwork"]}</b>}

                                <div className="row" >
                                    <div className="label">Local Net</div>
                                    <div className="edit"></div>
                                    <input className="input editable" onChange={e => UpdateInput("Networks", index, "LocalNetwork", e.target.value)} value={net.LocalNetwork} />
                                </div>
                                <div className="delete" onClick={() => RemoveField("Networks", index)} >Delete </div>

                            </div>

                            <div className="inner-seperator"></div>
                        </>
                    )
                })}


            </div>


            <div className="access panel">
                <div className="title">Access Control</div>
                <div className="add" onClick={() => AddAccessField()} >+</div>
                <div className="seperator"></div>


                {inputs["Access"].map((access, index) => {
                    return (<>

                        <div className={`access-card-${index}`} key={index}>

                            {access.Errors["Tag"] && <b className="error">{access.Errors["Tag"]}<br /></b>}
                            {access.Errors["UID"] && <b className="error">{access.Errors["UID"]}</b>}

                            <div className="row">
                                <div className="label">Tag</div>
                                <div className="edit"></div>
                                <input className="input editable" onChange={e => UpdateInput("Access", index, "Tag", e.target.value)} value={access.Tag} />

                                <div className="label label2">User ID</div>
                                <div className="edit edit2"></div>
                                <input className="input input2 editable" onChange={e => UpdateInput("Access", index, "UID", e.target.value)} value={access.UID} />

                                <div className="delete" onClick={() => RemoveField("Access", index)} >Delete </div>
                            </div>



                        </div>

                    </>)
                })}


            </div>
            {/* <div className="DNS panel">
                <div className="title">DNS Records</div>
                <div className="title">In progress..</div>
                <div className="seperator"></div>
            </div> */}
        </div >
    )

}

export default APCREATE;