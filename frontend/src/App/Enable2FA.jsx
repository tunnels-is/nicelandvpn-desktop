import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import QRCode from "react-qr-code";

import { ForwardToController, GetQRCode } from "../../wailsjs/go/main/Service";

import STORE from "../store";
import { EnvelopeClosedIcon, FrameIcon, LockClosedIcon } from "@radix-ui/react-icons";

const useForm = (props) => {
  const [inputs, setInputs] = useState({});
  const [errors, setErrors] = useState({});
  const [code, setCode] = useState({});
  const navigate = useNavigate();

  const HandleSubmit = async () => {

    let errors = {}
    let hasErrors = false

    if (!inputs["digits"]) {
      errors["digits"] = "Authenticator code missing"
      hasErrors = true
    } else {
      if (inputs["digits"].length < 6) {
        errors["digits"] = "Authenticator code is too short"
        hasErrors = true
      }
      if (inputs["digits"].length > 6) {
        errors["digits"] = "Authenticator code is too long"
        hasErrors = true
      }
    }

    if (!inputs["password"]) {
      errors["password"] = "Please enter your password"
      hasErrors = true
    }

    if (hasErrors) {
      setErrors({ ...errors })
      return
    }

    let user = STORE.Cache.GetObject("user")
    if (!user) {
      navigate("/loging")
    } else {
      inputs.Email = user.Email
    }

    let c = code.Value
    let firstSplit = c.split("&")
    let secondSPlit = firstSplit[1].split("=")
    let secret = secondSPlit[1]
    if (secret === "") {
      props.toggleError("Could not parse authenticator secret")
      setErrors({})
      return
    }

    inputs.Code = secret

    let FR = {
      Path: "v2/user/2fa/confirm",
      Method: "POST",
      JSONData: inputs,
      Timeout: 20000
    }


    props.toggleLoading({ logTag: "", tag: "2FA", show: true, msg: "Updating Two Factor Authentication", includeLogs: false })

    ForwardToController(FR).then((x) => {
      console.log("RETURN FOR QR")
      console.dir(x)
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code !== 200) {
          props.toggleError(x.Data)
        } else {
          let c = { ...code }
          c.Recovery = x.Data.Data
          setCode(c)
        }
      }

    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    setErrors({})
    props.toggleLoading(undefined)
  }

  const Get2FACode = async () => {

    let user = STORE.Cache.GetObject("user")
    if (!user) {
      navigate("/login")
    }

    let data = {
      Email: user.Email
    }

    GetQRCode(data).then((x) => {
      // x.Data.Recovery = "ALKJSDKLAJS ALKSJDKLAJS ASKLJDKALJSKL ASDKJAKLSJDKLA"
      setCode(x.Data)
    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    setErrors({})
  }

  const HandleInputChange = (event) => {
    setInputs(inputs => ({ ...inputs, [event.target.name]: event.target.value }));
  }

  return {
    inputs,
    HandleInputChange,
    HandleSubmit,
    errors,
    code,
    Get2FACode
  };
}


const Enable2FA = (props) => {

  const { inputs, HandleInputChange, HandleSubmit, errors, code, Get2FACode } = useForm(props);

  useEffect(() => {
    Get2FACode()
  }, [])


  return (
    <div className="two-factor-wrapper">

      {(code.Value && !code.Recovery) &&
        <div className="qr-wrapper">
          <div className="qr-code">
            <QRCode
              className="qr"
              style={{ height: "auto", maxWidth: "220px", width: "220px" }}
              value={code.Value}
              viewBox={`0 0 256 256`}
            ></QRCode>
          </div>
          <div className="text">
            Scan the code with a Two-Factor Authentication APP.<br />
            For Example: google authenticator or Aegis app
          </div>
        </div>
      }

      {code.Recovery &&
        <div className="recovery-codes">
          <div className="title">
            Please copy and store these recovery codes in a safe place, we recommend you write them down on paper. If you ever loose your authenticator app access you can use these codes to recover your account at a later date
          </div>
          <div className="codes">
            {code.Recovery}
          </div>
          <div className="notice">DO NOT STORE THESE CODES WITH YOUR PASSWORD</div>
        </div>
      }

      <div className="form" >

        <div className="input">
          <FrameIcon className="color-ok" width={40} height={30} center ></FrameIcon>
          <input className="code"
            placeholder={"Authenticator Code"} type="text"
            value={inputs["digits"]}
            name="digits"
            onChange={HandleInputChange} />
          {errors["digits"] && <div className="error">{errors["digits"]}</div>}
        </div>

        <div className="input">
          <LockClosedIcon className="color-ok" width={40} height={30} center ></LockClosedIcon>
          <input className="pass"
            placeholder={"Password"} type="password"
            value={inputs["password"]}
            name="password"
            onChange={HandleInputChange} />
          {errors["password"] !== "" && <div className="error">{errors["password"]}</div>}
        </div>

        <div className="input">
          <FrameIcon className="color-ok" width={40} height={30} center ></FrameIcon>
          <input className="recovery"
            placeholder="Recovery Code (optional)"
            type="text"
            value={inputs["recovery"]}
            name="recovery"
            onChange={HandleInputChange} />
        </div>


        <div className="buttons">
          <button className={`ok-button`}
            onClick={() => HandleSubmit()}
          >Confirm</button>
        </div>

      </div>

    </div >
  )
}

export default Enable2FA;