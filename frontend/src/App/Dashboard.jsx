import { useNavigate, Navigate } from "react-router-dom";
import React, { useState } from "react";

import toast from 'react-hot-toast';
import Loader from "react-spinners/ScaleLoader";

import { Connect, Switch } from "../../wailsjs/go/main/Service";

import STORE from "../store";
import { MagnifyingGlassIcon } from "@radix-ui/react-icons";


const Dashboard = (props) => {

  const [filter, setFilter] = useState("");
  const navigate = useNavigate();
  const [error, setError] = useState(undefined)

  const updateFilter = (event) => {
    setFilter(event.target.value)
  }

  const LogOut = () => {
    props.toggleError("You have been logged out")
    STORE.Cache.Clear()
  }

  const ConfirmQuickConnect = (country, ar) => {

    toast.success((t) =>
    (
      <div className="exit-confirm">
        <div className="text">
          Your are connecting to
        </div>
        <div className="text-big">
          {country}
        </div>
        <button className="exit" onClick={() => toast.dismiss(t.id)}>Cancel</button>
        <button className="cancel" onClick={() => {
          toast.dismiss(t.id)
          QuickConnectToVPN(country, ar)
        }
        }>Connect</button>
      </div>

    ), { id: "connect", duration: 999999 }
    )
  }

  const ConfirmConnect = (a, ar) => {

    toast.success((t) => (
      <div className="exit-confirm">
        <div className="text">
          Your are connecting to
        </div>
        <div className="text-big">
          {a.Tag}
        </div>
        <button className="cancel" onClick={() => {
          toast.dismiss(t.id)
          ConnectToVPN(a, ar)
        }
        }>Connect</button>
        <button className="exit" onClick={() => toast.dismiss(t.id)}>Cancel</button>
      </div>
    ), { id: "connect", duration: 999999 })

  }

  const QuickConnectToVPN = async (country, ar) => {
    if (!STORE.ActiveRouterSet(props.state)) {
      setError("No active router set, please wait a moment")
      return
    }

    props.toggleLoading({ logTag: "connect", tag: "CONNECT", show: true, msg: "Connecting you to a VPN in " + country, includeLogs: true })


    if (!user.DeviceToken) {
      LogOut()
      return
    }

    let method = Connect
    if (props.state?.Connected) {
      method = Switch
    }

    let CONNECT_FORM = {
      UserID: user._id,
      DeviceToken: user.DeviceToken.DT,
      Country: country,
      ROUTERID: ar.ROUTERID,
      GROUP: ar.GROUP,
    }

    method(CONNECT_FORM).then((x) => {
      if (x.Code === 401) {
        LogOut()
      }

      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code === 200) {
          STORE.Cache.Set("connected_quick", country)
          props.showSuccessToast("Connected", { Title: "CONNECTED", Body: "You have been connected to a country by code: " + country, TimeoutType: "default" })
        } else {
          props.toggleError(x.Data)
        }
      }

    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading(undefined)
  }

  const ConnectToVPN = async (a, ar) => {

    if (!STORE.ActiveRouterSet(props.state)) {
      return
    }

    props.toggleLoading({ logTag: "connect", tag: "CONNECT", show: true, msg: "Connecting you to VPN " + a.Tag, includeLogs: true })

    if (!user.DeviceToken) {
      LogOut()
      return
    }

    let method = undefined
    if (props.state?.ActiveAccessPoint) {
      method = Switch
    } else {
      method = Connect
    }

    // console.log("CONNECTING TO:", a.Tag, a.GROUP, a.ROUTERID)
    // console.log("SECOND ROUTER:", ar.GROUP, ar.ROUTERID)

    let ConnectForm = {
      UserID: user._id,
      DeviceToken: user.DeviceToken.DT,

      GROUP: ar.GROUP,
      ROUTERID: ar.ROUTERID,

      XGROUP: a.Router.GROUP,
      XROUTERID: a.Router.ROUTERID,
      DEVICEID: a.DEVICEID,
    }

    if (a.Networks) {
      ConnectForm.Networks = a.Networks
    }

    method(ConnectForm).then((x) => {
      if (x.Code === 401) {
        LogOut()
      }

      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code === 200) {

          STORE.Cache.Set("connected_quick", "XX")

          if (a.GEO) {
            props.showSuccessToast("Connected to VPN " + a.Tag + " @ " + a.GEO.CountryFull, { Title: "CONNECTED", Body: "Connected to VPN with IP: " + a.IP + " @ " + a.GEO.CountryFull, TimeoutType: "default" })
          } else {
            props.showSuccessToast("Connected to VPN " + a.Tag, { Title: "CONNECTED", Body: "Connected to VPN with IP: " + a.IP, TimeoutType: "default" })
          }

        } else {
          props.toggleError(x.Data)
        }
      }

    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading(undefined)

  }

  const NavigateToEditAP = (id) => {
    navigate("/accesspoint/" + id)
  }
  const NavigateToCreateAP = () => {
    navigate("create/accesspoint")
  }

  let user = STORE.GetUser()
  if (!user) {
    return (<Navigate to={"/login"} />)
  }

  const RenderServer = (ap, ar, editButton, isConnected) => {
    let method = undefined
    if (isConnected) {
      method = undefined
    } else {
      method = ConfirmConnect
    }

    if (!ap.Online) {
      return (
        <>
          <div className={`server`} onClick={() => NavigateToEditAP(ap._id)}>

            <div className="item tag tag-offline" >{ap.Tag} </div>
            <div className="item offline-text" >
              {`( OFFLINE )`}

            </div>
            <div className="item x3"></div>
            {/* <div className="item x3"></div> */}
            <div className="item x3"></div>
            <div className="item x3"> </div>
            <div className="item x3"></div>
            <div className="item x3"></div>
          </div>
        </>
      )

    }

    return (
      <>
        <div className={`server ${isConnected ? `is-connected` : ``}`} onClick={() => method(ap, ar)} >

          <div className="item tag"  >{ap.Tag}</div>

          <div className="item country" >
            {ap.GEO &&
              <>
                <img
                  className="country-flag"
                  src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/" + ap.GEO.Country.toLowerCase() + ".svg"}
                // src={"/src/assets/images/flag/" + ap.GEO.Country.toLowerCase() + ".svg"}
                />
                <div className="text">
                  {ap.GEO.Country}
                </div>
              </>
            }
            {!ap.GEO &&
              <img
                className="country-flag"
                src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/temp.svg"}
              />
            }
          </div>

          {ap.Router &&
            <>
              <div className="item x3">{ap.Router.Score}</div>
              <div className="item x3">{ap.Router.AvailableSlots} / {ap.Router.Slots}</div>
              <div className="item x3">{ap.Router.AvailableMbps / 1000}</div>
              <div className="item x3">{ap.Router.AIBP} / {ap.Router.AEBP}</div>
              <div className="item x3">{ap.Router.CPUP}</div>
              <div className="item x3">{ap.Router.RAMUsage}</div>
            </>
          }
        </div>
      </>
    )
  }

  let AccessPoints = []
  let Countries = []

  if (props.advancedMode) {

    if (props?.state?.AccessPoints) {

      if (filter && filter !== "") {

        props.state.AccessPoints.map(r => {

          let filterMatch = false
          if (r.Tag.includes(filter)) {
            filterMatch = true
          } else if (r.GEO.Country.includes(filter)) {
            filterMatch = true
          } else if (r.GEO.CountryFull.includes(filter)) {
            filterMatch = true
          }

          if (filterMatch) {
            AccessPoints.push(r)
          }

        })

      } else {
        AccessPoints = props.state.AccessPoints
      }

    }

  } else {
    if (props?.state?.AvailableCountries) {

      if (filter && filter !== "") {

        props.state.AvailableCountries.map(r => {

          let filterMatch = false
          if (r.includes(filter)) {
            filterMatch = true
          }

          if (filterMatch) {
            Countries.push(r)
          }

        })

      } else {
        Countries = props.state.AvailableCountries
      }

    }


  }

  let activeR = props.state?.ActiveRouter

  if (!props.advancedMode) {
    return (
      <div className="server-wrapper">

        <div className="search-wrapper">
          <MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
          <input type="text" className="search" onChange={updateFilter} placeholder="Search for Country Code .."></input>
        </div>

        {props.state?.ActiveAccessPoint &&
          <div className="simple-stats-bar" >
            <div className="simple-vpn">
              {`Connected to`} {props.state?.ActiveAccessPoint.Tag}
            </div>
          </div>
        }

        {Countries.length < 1 &&
          <Loader
            className="spinner"
            loading={true}
            color={"#20C997"}
            height={100}
            width={50}
          />
        }

        <div className="simple-list">
          {Countries.map(country => {
            return (
              <div className="item" onClick={() => ConfirmQuickConnect(country, activeR)}>

                <img
                  className="flag"
                  // src={"/src/assets/images/flag/" + country.toLowerCase() + ".svg"}
                  src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/" + country.toLowerCase() + ".svg"}
                />

                <div className="code">
                  {country}
                </div>

                <div className="code connect">
                  Connect
                </div>

              </div>)
          })}
        </div>

      </div >
    )
  }

  return (
    <div className="server-wrapper" >

      <div className="search-wrapper">
        <MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
        <input type="text" className="search" onChange={updateFilter} placeholder="Search for Tag or Country .."></input>
      </div>

      {activeR &&
        <div className="advanced-list advanced-list-bottom-margin" >

          <div className="header">
            <div className="title tag">Tag</div>
            <div className="title country">Country</div>
            <div className="title x3">QoS
              <span className="tooltiptext">{STORE.VPN_Tooltips[0]}</span>
            </div>
            <div className="title x3">Slots
              <span className="tooltiptext">{STORE.VPN_Tooltips[1]}</span>
            </div>
            <div className="title x3">Gbps
              <span className="tooltiptext">{STORE.VPN_Tooltips[5]}</span>
            </div>
            <div className="title x3">AB %
              <span className="tooltiptext">{STORE.VPN_Tooltips[2]}</span>
            </div>
            <div className="title x3">CPU %
              <span className="tooltiptext">{STORE.VPN_Tooltips[3]}</span>
            </div>
            <div className="title x3">RAM %
              <span className="tooltiptext">{STORE.VPN_Tooltips[4]}</span>
            </div>
          </div>

          {!props.state?.AccessPoints &&
            <Loader
              className="spinner"
              loading={true}
              color={"#20C997"}
              height={100}
              width={50}
            />
          }

          {props.state?.AccessPoints && props.state?.AccessPoints.length < 1 &&
            <Loader
              className="spinner"
              loading={true}
              color={"#20C997"}
              height={100}
              width={50}
            />
          }

          {props.state?.AccessPoints?.length > 0 &&
            AccessPoints.map(ap => {
              return RenderServer(ap, activeR, false, false)
            })
          }

        </div>
      }

    </div >
  );
}

export default Dashboard;