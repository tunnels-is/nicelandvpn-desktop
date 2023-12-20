import { useParams } from "react-router-dom";
import React, { useEffect, useState } from "react";
import STORE from "../store";
import dayjs from "dayjs";

const APFORM = (props) => {

    const customState = (props) => {
        const [inputs, setInputs] = useState({});
        const [dynamicCounter, setDynamicCounter] = useState({ Count: 0 });
        const [AP, setAP] = useState(undefined);
        const [changed, setChanged] = useState(false);

        const handleDelete = async (ID) => {
            console.log("REMOVING AP WITH ID", ID)
        }

        const handleSubmit = async (event) => {
            if (event) {
                event.preventDefault();
            }
            if (!changed) {
                props.toggleError("You have not made any changed to the VPN")
                return
            }


            console.log("SUBMITTING")
            console.dir(AP)
            console.dir(inputs)
            console.log("SUBMITTING")

            let user = STORE.Cache.GetObject("user")
            let form = {}
            let hasError = false

            form.UID = user._id
            form.DeviceToken = user.DeviceToken.DT

            form.Tag = inputs.Tag
            form.InternetAccess = inputs.InternetAccess
            form.LocalNetworkAccess = inputs.LocalNetworkAccess
            form.Access = []
            form.Networks = []
            form.CK = AP.ConnectKey
            form.RG = inputs.RG

            inputs["Access"].map(acc => {
                if (!acc.UID || acc.UID === "") {
                    hasError = true
                }
                if (!acc.Tag || acc.Tag === "") {
                    hasError = true
                }
                form.Access.push({ UID: acc.UID, Tag: acc.Tag })
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
                form.Networks.push({ Tag: net.Tag, Network: net.Network, LocalNetwork: net.LocalNetwork })
            })


            if (hasError) {
                props.toggleError("VPN can not be saved with errors")
                return
            }


            console.dir(form)

            let FR = {
                Path: "user/device/update",
                Method: "POST",
                JSONData: form,
                Timeout: 20000
            }

            try {
                // const response = await CLIENT.post("forward", JSON.stringify(FR));
                // console.dir(response)

                // let ad = STORE.Cache.GetObject("advanced-data-private")
                // ad.APS.map((ap, index) => {
                //     if (ad.APS[index].ConnectKey === AP.ConnectKey) {
                //         ad.APS[index].Tag = form.Tag
                //         ad.APS[index].Access = form.Access
                //         ad.APS[index].Networks = form.Networks
                //         ad.APS[index].InternetAccess = form.InternetAccess
                //         ad.APS[index].LocalNetworkAccess = form.LocalNetworkAccess
                //         ad.APS[index].Region = form.RG
                //     }
                // })
                // STORE.Cache.SetObject("advanced-data-private", ad)
                // props.showSuccessToast("VPN information updated")
            } catch (error) {
                props.toggleError(STORE.ParseResponseErrorMessage(error))
            }


        }

        const UpdateInput = (type, index, key, value) => {
            let i = { ...inputs }
            let hasErrors = false

            if (type === "InternetAccess") {
                console.log("UPDATING", type, i.InternetAccess)
                if (i.InternetAccess) {
                    i.InternetAccess = false
                } else {
                    i.InternetAccess = true
                }
            } else if (type === "LocalNetworkAccess") {
                console.log("UPDATING", type, i.LocalNetworkAccess)
                if (i.LocalNetworkAccess) {
                    i.LocalNetworkAccess = false
                } else {
                    i.LocalNetworkAccess = true
                }
            } else if (type === "Access") {
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
            } else if (type === "RG") {
                i.RG = value.toLowerCase()
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
            AP,
            setAP,
            RemoveField
        };
    }

    const { inputs, setInputs, dynamicCounter, setDynamicCounter, UpdateInput, handleSubmit, handleDelete, AP, setAP, RemoveField } = customState(props);


    let { id } = useParams()

    let GEOKeys = []
    let GEOExtraKeys = []
    let BasicKeys = []
    let StatsKeys = []

    useEffect(() => {
        let inputs = {}
        inputs.Access = []
        inputs.Networks = []
        let data = STORE.Cache.GetObject("advanced-data-private")
        let ldc = { ...dynamicCounter }
        console.log(id)
        let accesspoints = []
        accesspoints = data.D
        if (accesspoints !== undefined) {
            accesspoints.map(ap => {
                console.log("comparing APS", ap._id, id)
                if (ap._id === id) {
                    ap.Slots = ap.AvailableMbps / ap.UserMbps
                    // if (ap._id === id) {
                    setAP(ap)
                    inputs.Tag = ap.Tag
                    inputs.InternetAccess = ap.InternetAccess
                    inputs.LocalNetworkAccess = ap.LocalNetworkAccess


                    if (ap.Access) {
                        ap.Access.map(acc => {
                            ldc.Count = ldc.Count + 1
                            inputs.Access[ldc.Count] = {}
                            inputs.Access[ldc.Count].UID = acc.UID
                            inputs.Access[ldc.Count].Tag = acc.Tag
                            inputs.Access[ldc.Count].Errors = {}
                        })
                    }
                    if (ap.Networks) {
                        ap.Networks.map(n => {
                            ldc.Count = ldc.Count + 1
                            inputs.Networks[ldc.Count] = {}
                            inputs.Networks[ldc.Count].Network = n.Network
                            inputs.Networks[ldc.Count].LocalNetwork = n.LocalNetwork
                            inputs.Networks[ldc.Count].Tag = n.Tag
                            inputs.Networks[ldc.Count].Errors = {}
                        })
                    }
                }
            })

            setInputs(inputs)
            setDynamicCounter(ldc)
        }
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

    if (AP === undefined) {
        return (
            <h1>Could not find VPN</h1>
        )
    }

    console.dir(AP)
    Object.keys(AP).forEach((key) => {
        if (key === "GEO") {
            Object.keys(AP.GEO).forEach((geokey) => {
                GEOKeys.push(geokey)
            })
        } else if (key === "Access") {

        } else if (key === "RouterIP") {

        } else if (key === "Networks") {

        } else if (key === "Region") {

        } else if (key === "Tag") {

        } else if (key === "InternetAccess") {

        } else if (key === "LocalNetworkAccess") {

        } else if (key === "Updated" || key === "MS" || key === "UPTIME" || key === "UsedPercentage" || key === "CPUCores") {
            StatsKeys.push(key)
        } else if (key === "_id" || key === "DID" || key === "UID") {
            BasicKeys.push(key)
        } else {
            BasicKeys.push(key)
        }
    })

    // if (inputs["Access"]) {
    //     Object.keys(inputs).forEach(key => {
    //         AccessInputFields[key] = {}
    //         AccessInputFields[key].UID = inputs[key].UID
    //         AccessInputFields[key].Tag = inputs[key].Tag
    //         AccessInputFields[key].Errors = inputs[key].Errors
    //     })
    // }


    // console.dir(errors)
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



                {BasicKeys.map((key) => {
                    if (key === "UID") {
                        return (
                            <div className="row" key={key}>
                                <div className="label">Admin ID</div>
                                <input className="input" disabled value={AP[key]} />
                            </div>
                        )
                    } else {

                        return (
                            <div className="row" key={key}>
                                <div className="label">{key}</div>
                                <input className="input" disabled value={AP[key]} />
                            </div>
                        )
                    }
                })}


            </div>
            <div className="routing panel">
                <div className="title">GEO / Routing</div>
                <div className="seperator"></div>

                <div className="row">
                    <div className="label">Internet</div>
                    <div className="value toggle-button">
                        <label className="switch">
                            <input checked={inputs["InternetAccess"] ? true : false} type="checkbox" onClick={() => UpdateInput("InternetAccess", 0, "InternetAccess", null)} />
                            <span className="slider"></span>
                        </label>
                    </div>
                </div>

                <div className="row">
                    <div className="label">Local Network</div>
                    <div className="value toggle-button">
                        <label className="switch">
                            <input checked={inputs["LocalNetworkAccess"] ? true : false} type="checkbox" onClick={() => UpdateInput("LocalNetworkAccess", 0, "LocalNetworkAccess", null)} />
                            <span className="slider"></span>
                        </label>
                    </div>
                </div>



                {GEOKeys.map((key) => {
                    if (key === "Updated") {
                        return (
                            <div className="row" key={key}>
                                <div className="label">{key}</div>
                                <input className="input" disabled value=
                                    {dayjs(AP.GEO[key]).format("YYYY-MM-DD HH:mm:ss")}
                                />
                            </div>

                        )
                    } else if (key === "Country") {

                    } else if (key === "CountryFull") {
                        return (
                            <div className="row" key={key}>
                                <div className="label">Location</div>
                                <input className="input" disabled value={AP.GEO["Country"] + " / " + AP.GEO[key]} />
                            </div>
                        )

                    } else {
                        return (
                            <div className="row" key={key}>
                                <div className="label">{key}</div>
                                <input className="input" disabled value={AP[key]} />
                            </div>
                        )
                    }
                })}
                <div className="row" key={"geoextra"}>
                    <div className="label">Routing Tag</div>
                    {/* <input className="input" disabled value={AP["MID"].join("") + "-" + AP["ID"]} /> */}
                </div>
                {/* {GEOExtraKeys.map((key) => {
                    return (
                        <div className="row" key={key}>
                            <div className="label">{key}</div>
                            <input className="input" disabled value={AP[key]} />
                        </div>
                    )
                })} */}
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


                            {/* <div className="inner-seperator"></div> */}

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

export default APFORM;