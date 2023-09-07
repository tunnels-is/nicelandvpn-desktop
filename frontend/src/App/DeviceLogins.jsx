import React from "react";
import { useNavigate, Navigate } from "react-router-dom";

import dayjs from "dayjs";

import { ForwardToController } from "../../wailsjs/go/main/Service";

import STORE from "../store";

const DeviceLogins = (props) => {

  const navigate = useNavigate();

  let user = STORE.Cache.GetObject("user")
  if (user === undefined) {
    return (<Navigate replace to="/login" />)
  }

  const LogoutDevice = async (token) => {
    // props.toggleLoading({ logTag: "", tag: "LOGOUT", show: true, msg: "Logging out " + token.N, includeLogs: false })
    console.log("LOGGING OUT:", token)

    let user = STORE.Cache.GetObject("user")

    let LF = {
      DeviceToken: token.DT,
      Email: user.Email,
    }

    let FR = {
      Path: "v2/user/logout",
      Method: "POST",
      JSONData: LF,
      Timeout: 20000
    }

    ForwardToController(FR).then((x) => {
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (user.DeviceToken && user.DeviceToken.DT === token.DT) {
          STORE.Cache.DelObject("user");
          navigate("/login");
        } else {

          user.Tokens.map((t, index) => {
            if (t.DT === token.DT) {
              user.Tokens.splice(index, 1)
              props.showSuccessToast("Device ( " + t.N + " ) logged out")
            }
          })

          if (user.Tokens.length < 1) {
            STORE.CleanupOnLogout()
            return
          }

          STORE.Cache.SetObject("user", user)
        }
      }

    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    // props.toggleLoading(undefined)
  }

  let devices = user.Tokens.sort(function (a, b) {
    let bunix = dayjs(b.Created).unix()
    let aunix = dayjs(a.Created).unix()
    return bunix - aunix
  })

  return (
    <div className="device-wrapper">

      <div className="row header">
        <div className="item">
          Device Name
        </div>
        <div className="item">
          Date
        </div>
        <div className="item">
        </div>

      </div>

      {devices.map((t) => {
        let isCurrentLogin = false
        if (user?.DeviceToken?.DT === t.DT) {
          isCurrentLogin = true
        }
        return (

          <div className="row">

            {isCurrentLogin &&
              <div className="name item">
                <div className="current">{'(this device)'}</div>
                {t.N}
              </div>
            }
            {!isCurrentLogin &&
              <div className="name item">
                {t.N}
              </div>
            }

            <div className="date item">
              {dayjs(t.Created).format("DD-MM-YYYY HH:mm")}
            </div>

            <div className="logout item" onClick={() => LogoutDevice(t)}>
              Logout
            </div>


          </div>
        )

      })}
    </div >
  )

};

export default DeviceLogins;