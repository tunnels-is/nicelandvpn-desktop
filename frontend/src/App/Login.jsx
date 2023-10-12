import { useNavigate, Navigate } from "react-router-dom";
import React, { useEffect, useState } from "react";

import { v4 as uuidv4 } from 'uuid';
import { DesktopIcon, EnvelopeClosedIcon, ExclamationTriangleIcon, FrameIcon, HeartIcon, LockClosedIcon, QuestionMarkCircledIcon, Share1Icon } from "@radix-ui/react-icons";

import { ForwardToController } from "../../wailsjs/go/main/Service";

import STORE from "../store";

const useForm = (props) => {
  const [inputs, setInputs] = useState({})
  const [tokenLogin, setTokenLogin] = useState(false)
  const [errors, setErrors] = useState({})
  const navigate = useNavigate()
  const [mode, setMode] = useState(1)

  const RemoveToken = () => {
    setTokenLogin(false)
    errors["email"] = ""
    setErrors({ ...errors })
    setInputs(inputs => ({ ...inputs, ["email"]: "" }));
  }

  const GenerateToken = () => {
    let token = uuidv4()
    setTokenLogin(true)
    errors["email"] = "SAVE THIS TOKEN!"
    setErrors({ ...errors })
    setInputs(inputs => ({ ...inputs, ["email"]: token }));
  }

  const RegisterSubmit = async (event) => {

    let errors = {}
    let hasErrors = false

    if (!inputs["email"]) {
      errors["email"] = "Email / Token missing"
      hasErrors = true
    }

    if (inputs["email"]) {
      if (inputs["email"].length > 320) {
        errors["email"] = "Maximum 320 characters"
        hasErrors = true
      }

      if (!tokenLogin) {
        if (!inputs["email"].includes(".") || !inputs["email"].includes("@")) {
          errors["email"] = "Invalid email format"
          hasErrors = true

        }
      }
    }

    if (!inputs["password"]) {
      errors["password"] = "Password missing"
      hasErrors = true
    }
    if (!inputs["password2"]) {
      errors["password2"] = "Password confirm missing"
      hasErrors = true
    }

    if (inputs["password"] !== inputs["password2"]) {
      errors["password2"] = "Passwords do not match"
      hasErrors = true
    }

    if (inputs["password"]) {
      if (inputs["password"].length < 10) {
        errors["password"] = "Minimum 10 characters"
        hasErrors = true
      }
      if (inputs["password"].length > 255) {
        errors["password"] = "Maximum 255 characters"
        hasErrors = true
      }
    }

    if (hasErrors) {
      setErrors({ ...errors })
      return
    }

    let FR = {
      Path: "v2/user/create",
      Method: "POST",
      JSONData: inputs,
      Timeout: 20000
    }

    props.toggleLoading({ tag: "REGISTER", show: true, msg: "Creating your account" })

    ForwardToController(FR).then((x) => {
      if (x.Err) {
        props.toggleError(x.Err)
        setErrors({})
      } else {
        if (x.Code === 200) {
          STORE.Cache.Set("default-email", inputs["email"])
          props.showSuccessToast("Your account has been created, please log in", undefined)
          setErrors({})
          inputs["password"] = ""
          inputs["password2"] = ""
          setInputs({ ...inputs })
          setMode(1)
        } else {
          props.toggleError(x.Data)
          setErrors({})
        }
      }
    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading(undefined)
  }

  const HandleSubmit = async (event) => {

    let errors = {}
    let hasErrors = false

    if (!inputs["email"] || inputs["email"] === "") {
      errors["email"] = "Email / Token missing"
      hasErrors = true
    }

    if (!inputs["password"] || inputs["password"] === "") {
      errors["password"] = "Password missing"
      hasErrors = true
    }

    if (!inputs["devicename"] || inputs["devicename"] === "") {
      errors["devicename"] = "Device login name missing"
      hasErrors = true
    }

    if (mode === 2) {
      if (!inputs["digits"] || inputs["digits"] === "") {
        errors["digits"] = "Authenticator code missing"
        hasErrors = true
      }

      if (inputs["digits"] && inputs["digits"].length < 6) {
        errors["digits"] = "Code needs to be at least 6 digits"
        hasErrors = true
      }
    }

    if (mode === 3) {
      if (!inputs["recovery"] || inputs["recovery"] === "") {
        errors["recovery"] = "Recovery code missing"
        hasErrors = true
      }
    }

    if (hasErrors) {
      setErrors({ ...errors })
      return
    }

    props.toggleLoading({ tag: "LOGIN", show: true, msg: "Logging you in..." })

    let token = STORE.Cache.Get(inputs["email"] + "_" + "TOKEN")

    if (token !== null) {
      inputs.DeviceToken = token
    }

    // let version = STORE.Cache.Get("version")
    inputs["version"] = props.state.Version
    let FR = {
      Path: "v2/user/login",
      Method: "POST",
      JSONData: inputs,
      Timeout: 20000
    }

    STORE.Cache.Set("default-device-name", inputs["devicename"])
    STORE.Cache.Set("default-email", inputs["email"])

    ForwardToController(FR).then((x) => {
      console.dir(x)
      if (x.Err) {
        props.toggleError(x.Err)
        STORE.Cache.DelObject("user")
      } else {
        if (x.Code === 200) {
          STORE.Cache.Set(inputs["email"] + "_" + "TOKEN", x.Data.DeviceToken.DT)
          STORE.Cache.SetObject("user", x.Data);
          props.toggleLoading(undefined)
          navigate("/")
        } else {
          props.toggleError(x.Data)
          setErrors({})
        }
      }
    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading(undefined)

  }

  const ResetSubmit = async () => {

    let errors = {}
    let hasErrors = false

    if (!inputs["email"]) {
      errors["email"] = "Email / Token missing"
      hasErrors = true
    }

    if (inputs["email"]) {
      if (!inputs["email"].includes(".") || !inputs["email"].includes("@")) {
        errors["email"] = "Email address format is incorrect"
        hasErrors = true
      }
    }

    if (!inputs["password"]) {
      errors["password"] = "Password missing"
      hasErrors = true
    }

    if (inputs["password"] && inputs["password"].length < 9) {
      errors["password"] = "Password needs to be at least 9 characters in length"
      hasErrors = true
    }

    if (inputs["password"] && inputs["password"].length > 255) {
      errors["password"] = "Password can not be longer then 255 characters"
      hasErrors = true
    }

    if (!inputs["password2"]) {
      errors["password2"] = "Password confirmation missing"
      hasErrors = true
    }

    if (inputs["password"] !== inputs["password2"]) {
      errors["password"] = "Passwords do not match"
      hasErrors = true
    }


    if (!inputs["code"]) {
      errors["code"] = "code missing"
      hasErrors = true
    }

    if (hasErrors) {
      setErrors({ ...errors })
      return
    }

    let request = {
      Email: inputs["email"],
      NewPassword: inputs["password"],
      ResetCode: inputs["code"],
    }

    props.toggleLoading({ tag: "RESET", show: true, msg: "Changing your password, please wait a moment..." })

    let FR = {
      Path: "user/reset/password",
      Method: "POST",
      JSONData: request,
      Timeout: 20000
    }

    ForwardToController(FR).then((x) => {
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code === 200) {
          setErrors({})
          props.showSuccessToast("Password has been changed", undefined)
          inputs["password"] = ""
          inputs["password2"] = ""
          inputs["code"] = ""
          setInputs({ ...inputs })
          setMode(1)
        } else {
          setErrors({})
          props.toggleError("Unable to reset password, please try again in a few seconds..")
        }
      }

    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading({ tag: "RESET", show: false })
  }


  const GetCode = async (event) => {

    let errors = {}
    let hasErrors = false


    if (!inputs["email"]) {
      errors["email"] = "Email missing"
      hasErrors = true
    }

    if (inputs["email"]) {
      if (!inputs["email"].includes(".") || !inputs["email"].includes("@")) {
        errors["email"] = "Email address format is incorrect"
        hasErrors = true

      }
    }

    if (hasErrors) {
      setErrors({ ...errors })
      return
    }

    let request = {
      Email: inputs["email"],
    }

    let FR = {
      Path: "user/reset/code",
      Method: "POST",
      JSONData: request,
      Timeout: 20000
    }

    props.toggleLoading({ tag: "RESET", show: true, msg: "Requesting reset code, please wait..." })

    ForwardToController(FR).then((x) => {
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code === 200) {
          props.showSuccessToast("Password reset code has been sent to your email inbox!")
        } else {
          props.toggleError("Unable to request password reset code, please try again in a few seconds..")
        }
      }

    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading({ tag: "RESET", show: false })
  }


  const HandleInputChange = (event) => {
    setInputs(inputs => ({ ...inputs, [event.target.name]: event.target.value }));
  }

  const ManualInputChange = (key, value) => {
    setInputs(inputs => ({ ...inputs, [key]: value }));
  }

  return {
    inputs,
    setInputs,
    HandleInputChange,
    ManualInputChange,
    HandleSubmit,
    errors,
    navigate,
    setMode,
    mode,
    RegisterSubmit,
    GenerateToken,
    tokenLogin,
    ResetSubmit,
    GetCode,
    RemoveToken,
  };
}

const Login = (props) => {

  const {
    inputs,
    setInputs,
    HandleInputChange,
    HandleSubmit,
    errors,
    navigate,
    setMode,
    mode,
    RegisterSubmit,
    GenerateToken,
    tokenLogin,
    ResetSubmit,
    GetCode,
    RemoveToken,
  } = useForm(props);

  const GetDefaults = () => {
    let i = { ...inputs }

    let defaultDeviceName = STORE.Cache.Get("default-device-name")
    if (defaultDeviceName) {
      i["devicename"] = defaultDeviceName
    }

    let defaultEmail = STORE.Cache.Get("default-email")
    if (defaultEmail) {
      i["email"] = defaultEmail
    }

    setInputs(i)
  }

  useEffect(() => {

    // window.addEventListener("keypress", function (event) {

    //   if (event.key === "Enter") {
    //     // event.preventDefault()
    //     if (mode === 1) {
    //       HandleSubmit()
    //     } else if (mode === 2) {
    //       RegisterSubmit()
    //     } else if (mode === 3) {
    //       HandleSubmit()
    //     } else if (mode === 4) {
    //       ResetSubmit()
    //     }
    //   }
    // });


    GetDefaults()
  }, [])

  const EmailInput = () => {
    return (
      <div className="input">
        <EnvelopeClosedIcon className="color-ok" width={40} height={30} center ></EnvelopeClosedIcon>
        <input
          className="email-input"
          autocomplete="off"
          type="email"
          placeholder={"Email / Token"}
          value={inputs["email"]}
          name="email"
          onChange={HandleInputChange} />
        {errors["email"] !== "" &&
          <div className="error">
            {errors["email"]}
          </div>
        }
      </div>
    )
  }

  const DeviceInput = () => {
    return (
      <div className="input">
        <DesktopIcon className="color-ok" width={40} height={30} center ></DesktopIcon>
        <input className="device-input"
          type="text"
          placeholder={"Device Name"}
          value={inputs["devicename"]}
          name="devicename"
          onChange={HandleInputChange} />
        {errors["devicename"] &&
          <div className="error">
            {errors["devicename"]}
          </div>
        }
      </div>
    )

  }

  const PasswordInput = () => {
    return (
      <div className="input">
        <LockClosedIcon className="color-ok" width={40} height={30} center ></LockClosedIcon>
        <input className=" pass-input"
          type="password"
          placeholder={"Password"}
          value={inputs["password"]}
          name="password"
          onChange={HandleInputChange} />
        {errors["password"] &&
          <div className="error">
            {errors["password"]}
          </div>
        }
      </div>
    )
  }

  const TwoFactorInput = () => {
    return (
      <div className="input">
        <LockClosedIcon className="color-ok" width={40} height={30} center ></LockClosedIcon>
        <input className=" code-input"
          type="text"
          placeholder={"Authenticator Code (optional)"}
          value={inputs["digits"]}
          name="digits"
          onChange={HandleInputChange} />
        {errors["digits"] &&
          <div className="error">
            {errors["digits"]}
          </div>
        }
      </div>
    )
  }

  const ConfirmPasswordInput = () => {
    return (
      <div className="input">
        <LockClosedIcon className="color-ok" width={40} height={30} center ></LockClosedIcon>
        <input className="code-input"
          type="password"
          placeholder={"Confirm Password"}
          value={inputs["password2"]}
          name="password2"
          onChange={HandleInputChange} />
        {errors["password2"] &&
          <div className="error">
            {errors["password2"]}
          </div>
        }
      </div>
    )
  }

  const TokenInput = () => {
    return (
      <div className="input">
        <FrameIcon className="color-ok" width={40} height={30} center ></FrameIcon>
        <input className=" token-input"
          autocomplete="off"
          placeholder={"Token / Token"}
          type="text"
          value={inputs["email"]}
          name="email"
          onChange={HandleInputChange} />
        {errors["email"] &&
          <div className="error">
            {errors["email"]}
          </div>
        }
      </div>
    )
  }

  const AffiliateInput = () => {
    return (
      <div className="input">
        <Share1Icon className="color-ok" width={40} height={30} center ></Share1Icon>
        <input className=" code-input"
          type="text"
          placeholder={"Affiliate Code (optional)"}
          value={inputs["code"]}
          name="code"
          onChange={HandleInputChange} />
      </div>
    )
  }

  const CodeInput = () => {
    return (
      <div className="input">
        <FrameIcon className="color-ok" width={40} height={30} center ></FrameIcon>
        <input
          className="code-input"
          autocomplete="off"
          type="text"
          placeholder={"Reset Code"}
          // value={inputs["email"]}
          name="code"
          onChange={HandleInputChange} />
        {errors["code"] &&
          <div className="error">
            {errors["code"]}
          </div>
        }
      </div>

    )
  }

  const RecoveryInput = () => {
    return (
      <div className="input">
        <FrameIcon className="color-ok" width={40} height={30} center ></FrameIcon>
        <input className=" recovery-input"
          type="text"
          placeholder={"Two Factor Recovery Code"}
          value={inputs["recovery"]}
          name="recovery"
          onChange={HandleInputChange} />
        {errors["recovery"] &&
          <div className="error">
            {errors["recovery"]}
          </div>
        }
      </div>

    )
  }

  const LoginForm = () => {
    return (
      <div className="form" >

        {EmailInput()}
        {DeviceInput()}
        {PasswordInput()}
        {TwoFactorInput()}

        <div className="buttons">
          <button className={`ok-button`}
            onClick={HandleSubmit} >Login</button>
        </div>

      </div>
    )
  }
  const RegisterAnonForm = () => {
    return (
      <div className="form">
        <div className="warning">Save your login token in a secure place, it is the only form of authentication you have for your account. If you loose the token your account is lost forever.</div>

        {TokenInput()}
        {PasswordInput()}
        {ConfirmPasswordInput()}
        {AffiliateInput()}

        <div className="buttons">

          <button className={`ok-button`}
            onClick={RegisterSubmit}
          >Register</button>


        </div>

      </div>)
  }


  const RegisterForm = () => {
    return (
      <div className="form">

        {tokenLogin &&
          TokenInput()
        }

        {!tokenLogin &&
          EmailInput()
        }

        {PasswordInput()}
        {ConfirmPasswordInput()}
        {AffiliateInput()}

        <div className="buttons">

          <button className={`ok-button`}
            onClick={RegisterSubmit}
          >Register</button>


        </div>

      </div>)
  }

  const ResetPasswordForm = () => {
    return (
      <div className="form" >

        <div className="input">
          <div className="get-code-button" onClick={() => GetCode()}>Get Code In Email</div>
        </div>

        {EmailInput()}
        {PasswordInput()}
        {ConfirmPasswordInput()}
        {CodeInput()}

        <div className="buttons">
          <button className={`ok-button`}
            onClick={() => ResetSubmit()} >Reset Password</button>
        </div>

      </div>
    )
  }


  const RecoverTwoFactorForm = () => {
    return (
      <div className="form" >

        {EmailInput()}
        {PasswordInput()}
        {RecoveryInput()}

        <div className="buttons">
          <button className={`ok-button`}
            onClick={HandleSubmit} >Login</button>
        </div>

      </div>
    )
  }

  return (
    <div className="login-wrapper">

      {mode === 1 &&
        LoginForm()
      }
      {mode === 2 &&
        RegisterForm()
      }
      {mode === 4 &&
        ResetPasswordForm()
      }
      {mode === 3 &&
        RecoverTwoFactorForm()
      }
      {mode === 5 &&
        RegisterAnonForm()
      }

      <div className="login-options">

        <button className="button"
          onClick={() => {
            RemoveToken()
            setMode(2)
          }} >Register</button>

        <button className="button"
          onClick={() => {
            GenerateToken()
            setMode(5)
          }} >Register Anonymously</button>

        <button className="button"
          onClick={() => setMode(1)} >Login</button>


        <button className="button"
          onClick={() => setMode(4)} >Reset Password</button>

        <button className="button"
          onClick={() => setMode(3)} >2FA Recovery</button>

      </div>

    </div>
  )

}

export default Login;