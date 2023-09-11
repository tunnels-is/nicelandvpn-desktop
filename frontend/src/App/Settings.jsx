import React, { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";

import { v4 as uuidv4 } from 'uuid';
import dayjs from "dayjs";
import { DesktopIcon, FileTextIcon, LockClosedIcon, Pencil2Icon, PersonIcon } from '@radix-ui/react-icons'

import { ForwardToController, ResetEverything, SetConfig, OpenFileDialogForRouterFile, EnableBlocklist, DisableBlocklist, RebuildDomainBlocklist, EnableAllBlocklists, DisableAllBlocklists } from "../../wailsjs/go/main/Service";
import { CloseApp, OpenURL } from "../../wailsjs/go/main/App";

import STORE from "../store";


const useForm = (props) => {

  const [inputs, setInputs] = useState({});
  const [user, setUser] = useState();
  const [autoLogout, setAutoLogout] = useState(STORE.Cache.GetBool("auto-logout"))
  const navigate = useNavigate();
  const [blockList, setBlockList] = useState([])

  // const [capture, setCapture] = useState(false)
  // const [whitelist, setWhitelist] = useState(true)

  const ToggleAllBlocking = (enabled) => {
    console.log("TOGGLE ALL")

    if (enabled) {
      EnableAllBlocklists()

      blockList.forEach((item, index) => {
        blockList[index].Enabled = true
      })

    } else {
      DisableAllBlocklists()

      blockList.forEach((item, index) => {
        blockList[index].Enabled = false
      })

    }

    setBlockList([...blockList])
  }


  const ApplyBlocklistConfigurations = async () => {
    props.toggleLoading({ logTag: "", tag: "DOMAIN-BLOCK", show: true, msg: "Creating a new combined blocklist..", includeLogs: false })
    try {
      await RebuildDomainBlocklist()
    } catch (error) {
      console.dir(error)
    }
    props.toggleLoading(undefined)
  }

  const ToggleBlockList = async (tag) => {

    blockList.forEach((item, index) => {
      if (item.Tag === tag) {
        if (item.Enabled) {
          DisableBlocklist(tag)
          blockList[index].Enabled = false
        } else {
          EnableBlocklist(tag)
          blockList[index].Enabled = true
        }
      }
    })

    setBlockList([...blockList])
  }


  // const EnableWhitelist = () => {
  //   setWhitelist(true)
  //   EnableDNSWhitelist()
  // }

  // const DisableWhitelist = () => {
  //   setWhitelist(false)
  //   DisableDNSWhitelist()
  // }

  // const StartCapture = () => {
  //   setCapture(true)
  //   StartDNSCapture()
  // }

  // const StopCapture = () => {
  //   setCapture(false)
  //   StopDNSCapture()
  // }


  const ToggleAutoLogout = () => {
    let AL = STORE.Cache.GetBool("auto-logout")
    if (AL) {
      AL = false
    } else {
      AL = true
    }
    setAutoLogout(AL)
    STORE.Cache.Set("auto-logout", AL)
  }

  const Reset = async () => {

    await props.disconnectFromVPN()

    props.toggleLoading({ logTag: "", tag: "NETRESET", show: true, msg: "Recovering network settings", includeLogs: false })

    ResetEverything().then((x) => {
      props.showSuccessToast("Network Settings have been reset", undefined)
    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading(undefined)
  }

  const UpdateRouterFile = async (clearFile) => {

    props.toggleLoading({ logTag: "loader", tag: "ROUTERFILE", show: true, msg: "Updating routers from file", includeLogs: true })

    await OpenFileDialogForRouterFile(clearFile).then((x) => {

    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading(undefined)
  }

  const UpdateUser = async () => {

    // try {
    props.toggleLoading({ logTag: "", tag: "USER-UPDATE", show: true, msg: "Updating User Settings", includeLogs: false })


    let FORM = {
      Email: user.Email,
      DeviceToken: user.DeviceToken.DT,
      APIKey: inputs["APIKey"].APIKey
    }

    if (!FORM.APIKey || FORM.APIKey === "") {
      props.toggleError("API Key missing, please generate a new API Key.")
      props.toggleLoading(undefined)
      return
    }

    let FR = {
      Path: "v2/user/update",
      Method: "POST",
      JSONData: FORM,
      Timeout: 20000
    }

    ForwardToController(FR).then((x) => {
      console.dir(x)
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code === 200) {
          let nu = { ...user }
          nu.APIKey = FORM.APIKey
          STORE.Cache.SetObject("user", nu)
          props.showSuccessToast("User updated", undefined)
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

  const UpdateConfig = async () => {

    props.toggleLoading({ logTag: "loader", tag: "CONFIG", show: true, msg: "Updating configurations", includeLogs: false })

    let config = STORE.Cache.GetObject("config")
    let newConfig = { ...config }

    newConfig.DNS1 = inputs["DNS"].DNS1
    newConfig.DNS2 = inputs["DNS"].DNS2
    newConfig.RouterFilePath = inputs["RP"].RP
    newConfig.DebugLogging = inputs["DB"].DB
    newConfig.AutoReconnect = inputs["AR"].AR
    newConfig.KillSwitch = inputs["KS"].KS
    newConfig.DisableIPv6OnConnect = inputs["IP6"].IP6
    newConfig.CloseConnectionsOnConnect = inputs["CC"].CC

    SetConfig(newConfig).then((x) => {
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        STORE.Cache.SetObject("config", newConfig)
        props.showSuccessToast("Config saved", undefined)
        STORE.Cache.Set("aps-timeout", dayjs().subtract(120, "s").unix())
      }
    }).catch((e) => {
      console.dir(e)
      props.toggleError("Unknown error, please try again in a moment")
    })

    props.toggleLoading(undefined)
  }

  const ToggleSubscription = async (toggle) => {

    props.toggleLoading({ logTag: "", tag: "DISABLE", show: true, msg: "Updating your subscription", includeLogs: false })

    let user = STORE.Cache.GetObject("user")

    let FR = {
      Method: "POST",
      Timeout: 20000,
      Path: "v2/user/toggle/substatus",
      JSONData: {
        DeviceToken: user.DeviceToken.DT,
        Email: user.Email,
        Disable: toggle
      }
    }

    ForwardToController(FR).then((x) => {
      console.dir(x)
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code === 200) {
          props.showSuccessToast("Subscription status updated", undefined)
          user.CancelSub = toggle
          STORE.Cache.SetObject("user", user)
          setUser(user)
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

  const DisableAccount = async () => {


    props.toggleLoading({ logTag: "", tag: "DISABLE", show: true, msg: "Disabling your account", includeLogs: false })

    let user = STORE.Cache.GetObject("user")

    let FR = {
      Method: "POST",
      Timeout: 20000,
      Path: "v2/user/toggle/status",
      JSONData: {
        DeviceToken: user.DeviceToken.DT,
        ID: user._id,
        Disabled: true
      }
    }

    ForwardToController(FR).then((x) => {
      console.dir(x)
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code === 200) {
          props.showSuccessToast("Account Disabled", undefined)
          STORE.Cache.Clear()
          navigate("/login")
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

  const EnableAccount = async () => {
    props.toggleLoading({ logTag: "", tag: "ENABLE", show: true, msg: "Enabling your account", includeLogs: false })

    let user = STORE.Cache.GetObject("user")

    let FR = {
      Method: "POST",
      Timeout: 20000,
      Path: "v2/user/toggle/status",
      JSONData: {
        DeviceToken: user.DeviceToken.DT,
        ID: user._id,
        Disabled: false
      }
    }

    ForwardToController(FR).then((x) => {
      console.dir(x)
      if (x.Err) {
        props.toggleError(x.Err)
      } else {
        if (x.Code === 200) {
          props.showSuccessToast("Account Enabled", undefined)
          user.Disabled = false
          STORE.Cache.SetObject("user", user)
          setUser(user)
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

  const ToggleCC = () => {
    let i = { ...inputs }
    if (i["CC"].CC === true) {
      i["CC"].CC = false
    } else {
      i["CC"].CC = true
    }
    setInputs(i)
  }

  const ToggleLogging = () => {
    let i = { ...inputs }
    if (i["DB"].DB === true) {
      i["DB"].DB = false
    } else {
      i["DB"].DB = true
    }
    setInputs(i)
  }

  const ToggleKS = () => {
    let i = { ...inputs }
    if (i["KS"].KS === true) {
      i["KS"].KS = false
    } else {
      i["KS"].KS = true
    }
    setInputs(i)
  }

  const ToggleIP6 = () => {
    let i = { ...inputs }
    if (i["IP6"].IP6 === true) {
      i["IP6"].IP6 = false
    } else {
      i["IP6"].IP6 = true
    }
    setInputs(i)
  }

  const ToggleAR = () => {
    let i = { ...inputs }
    if (i["AR"].AR === true) {
      i["AR"].AR = false
    } else {
      i["AR"].AR = true
    }
    setInputs(i)
  }

  const UpdateInput = (index, key, value) => {

    let i = { ...inputs }
    let hasErrors = false

    i[index][key] = value

    let skipErrorValidation = false
    if (index !== "APIKey") {
      skipErrorValidation = true
    }


    if (!skipErrorValidation) {
      if (value === "") {
        i[index].Errors[key] = key + " missing"
      } else {
        i[index].Errors[key] = ""

        if (index === "Email") {
          if (!value.includes("@") || !value.includes(".")) {
            i[index].Errors[key] = key + " format is wrong"
          }
        }
      }
    }

    setInputs(i)

    if (hasErrors) {
      props.setSuccess("")
      props.toggleError("Error in settings input")
    }

  }

  return {
    inputs,
    setInputs,
    user,
    setUser,
    UpdateInput,
    UpdateConfig,
    Reset,
    DisableAccount,
    EnableAccount,
    UpdateRouterFile,
    ToggleLogging,
    ToggleAR,
    ToggleKS,
    UpdateUser,
    ToggleSubscription,
    autoLogout,
    ToggleAutoLogout,
    ToggleBlockList,
    ApplyBlocklistConfigurations,
    setBlockList,
    blockList,
    ToggleIP6,
    ToggleAllBlocking,
    ToggleCC,
  }
}

const Settings = (props) => {

  const navigate = useNavigate();

  const { inputs, setInputs, user, setUser, UpdateInput, UpdateConfig, Reset, DisableAccount, EnableAccount, UpdateRouterFile, ToggleLogging, ToggleAR, ToggleKS, UpdateUser, ToggleSubscription, autoLogout, ToggleAutoLogout, ToggleBlockList, ApplyBlocklistConfigurations, setBlockList, blockList, ToggleIP6, ToggleAllBlocking, ToggleCC } = useForm(props);

  const inputFile = useRef(null)

  const ResetApp = async () => {
    await props.disconnectFromVPN()
    STORE.Cache.Clear()
    CloseApp()
  }

  const CreateApiKey = () => {
    let i = { ...inputs }
    i["APIKey"].APIKey = uuidv4()
    setInputs(i)
  }

  useEffect(() => {

    if (props.state?.BlockLists?.length > 0) {
      setBlockList(props.state.BlockLists)
    }

    let i = {}
    i["Email"] = {}
    i["Email"].Errors = {}

    i["Password"] = {}
    i["Password"].Errors = {}

    i["APIKey"] = {}
    i["APIKey"].Errors = {}

    let u = STORE.Cache.GetObject("user");
    let newUser = undefined
    if (u) {
      newUser = { ...u }
    }

    if (newUser) {
      i["Email"].Email = newUser.Email
      i["Password"].Password = newUser.Password
      i["APIKey"].APIKey = newUser.APIKey
      let splitDate = u.SubExpiration.split("T")
      let finalDate = dayjs(splitDate[0]).format("DD MMMM YYYY")
      newUser.SubExpirationString = finalDate
      setUser(newUser)
    }

    i["RP"] = {}
    i["RP"].Errors = {}

    i["DB"] = {}
    i["DB"].Errors = {}

    i["KS"] = {}
    i["KS"].Errors = {}

    i["CC"] = {}
    i["CC"].Errors = {}

    i["IP6"] = {}
    i["IP6"].Errors = {}

    i["AR"] = {}
    i["AR"].Errors = {}

    i["DNS"] = {}
    i["DNS"].Errors = {}
    i["DNS"].DNS1 = ""
    i["DNS"].DNS2 = ""

    let config = STORE.Cache.GetObject("config")
    if (config) {
      i["DNS"].DNS1 = config.DNS1
      i["DNS"].DNS2 = config.DNS2
      i["RP"].RP = config.RouterFilePath
      i["DB"].DB = config.DebugLogging
      i["KS"].KS = config.KillSwitch
      i["AR"].AR = config.AutoReconnect
      i["IP6"].IP6 = config.DisableIPv6OnConnect
      i["CC"].CC = config.CloseConnectionsOnConnect
    } else {
      i["DNS"].DNS0 = "1.1.1.1"
      i["DNS"].DNS2 = "8.8.8.8"
      i["DB"].DB = true
    }

    setInputs(i)


  }, []);

  const NavigateToDebug = () => {
    navigate("/debug")
  }
  const NavigateToSubscriptions = () => {
    OpenURL("https://www.nicelandvpn.is/#/pricing")
  }

  const NavigateTo2fa = () => {
    navigate("/twofactor")
  }

  const NavigateToLogs = () => {
    navigate("/logs")
  }
  const NavigateToTokens = () => {
    navigate("/tokens")
  }

  return (
    <div className="settings-wrapper">
      {user &&
        <>
          <div className="panel">

            <div className="header">
              <PersonIcon className="icon"></PersonIcon>
              <div className="title account">Account</div>
            </div>

            <div className="seperator"></div>

            <div className="item">
              <div className="title email-color">{inputs["Email"].Email}</div>
            </div>

            <div className="item">
              <div className="title">Account Status</div>
              <div className="value">{`${user.Disabled ? "Disabled" : "Active"}`}</div>
            </div>

            <div className="item">
              <div className="title">Last Account Update</div>
              <div className="value">
                {dayjs(user.Updated).format("DD-MM-YYYY HH:mm:ss")}
              </div>
            </div>

            <div className="item extra-space">
              <div className="title">Subcription</div>
              <div className="value">
                {user.SubLevel === 1 &&
                  <>Nice</>
                }
                {user.SubLevel === 2 &&
                  <>Nicer</>
                }
                {user.SubLevel === 3 &&
                  <>Nicest</>
                }
                {(user.SubLevel === 0 || user.SubLevel === 666) &&
                  <div className="link-color neutral-color" onClick={() => NavigateToSubscriptions()}>Click To Subscribe</div>
                }
              </div>
            </div>

            {(user.SubLevel > 0 && user.SubLevel !== 666) &&
              <>
                <div className="item">
                  <div className="title">Status</div>
                  <div className="value">{`${user.CancelSub ? "Disabled" : "Active"}`}</div>
                </div>
                <div className="item">
                  <div className="title">Expires</div>
                  <div className="value">{user.SubExpirationString}</div>
                </div>
                {user.CancelSub &&
                  <div className="item">
                    <div className="title neutral-color" onClick={() => ToggleSubscription(false)} >Enable Subscription</div>
                  </div>
                }
                {!user.CancelSub &&
                  <div className="item">
                    <div className="title warning-color" onClick={() => ToggleSubscription(true)} >Disable Subscription</div>
                  </div>
                }
              </>
            }
            {!user?.Disabled &&
              <div className="item">
                <div className="title warning-color" onClick={() => DisableAccount()} >Disable Account</div>
              </div>
            }


            <div className="item extra-space">
              <div className="title neutral-color" onClick={() => NavigateToTokens()} >Device Logins</div>
            </div>
            <div className="item">
              <div className="title  neutral-color" onClick={() => NavigateTo2fa()} >Two-Factor Authentication</div>
            </div>

            {user?.Disabled &&
              <div className="item">
                <div className="title ok-color" onClick={() => EnableAccount()} >Enable Account</div>
              </div>
            }


          </div>

        </>

      }



      <div className="panel">

        <div className="header">
          <DesktopIcon className="icon"></DesktopIcon>
          <div className="title">Utility</div>
          <div className="save-config title neutral-color" onClick={() => UpdateConfig()}>Save</div>
        </div>

        <div className="seperator"></div>

        <div className="item">
          <div className="am toggle-button">
            <label className="switch">
              <input checked={props.advancedMode ? true : false} type="checkbox" onClick={() => props.toggleAdvancedMode()} />
              <span className="slider"></span>
            </label>
            <div className="text">
              Advanced Mode
            </div>
          </div>
        </div>

        <div className="item less-space">
          <div className="am toggle-button">
            <label className="switch">
              <input checked={(inputs["AR"] && inputs["AR"].AR) ? true : false} type="checkbox" onClick={() => ToggleAR()} />
              <span className="slider"></span>
            </label>
            <div className="text">
              Auto Reconnect
            </div>
          </div>
        </div>

        {props.advancedMode &&
          <>

            <div className="item less-space">
              <div className="am toggle-button">
                <label className="switch">
                  <input checked={(inputs["KS"] && inputs["KS"].KS) ? true : false} type="checkbox" onClick={() => ToggleKS()} />
                  <span className="slider"></span>
                </label>
                <div className="text">
                  Killswitch
                </div>
              </div>
            </div>


            <div className="item less-space">
              <div className="am toggle-button">
                <label className="switch">
                  <input checked={(inputs["IP6"] && inputs["IP6"].IP6) ? true : false} type="checkbox" onClick={() => ToggleIP6()} />
                  <span className="slider"></span>
                </label>
                <div className="text">
                  Disable IPv6
                </div>
              </div>
            </div>

            <div className="item less-space">
              <div className="am toggle-button">
                <label className="switch">
                  <input checked={(inputs["DB"] && inputs["DB"].DB) ? true : false} type="checkbox" onClick={() => ToggleLogging()} />
                  <span className="slider"></span>
                </label>
                <div className="text">
                  Logging
                </div>
              </div>
            </div>
          </>

        }






        <div className="item extra-space">
          <div className="title neutral-color" onClick={() => Reset()} >Reset Network Settings</div>
        </div>

        <div className="item">
          <div className="title warning-color" onClick={() => ResetApp()} >Reset App Settings</div>
        </div>

      </div>

      {props.advancedMode &&
        <div className="panel">

          <div className="header">
            <FileTextIcon className="icon"></FileTextIcon>
            <div className="title">Routing</div>
            <div className="save-config title neutral-color" onClick={() => UpdateConfig()}>Save</div>
          </div>

          <div className="seperator"></div>

          <div className="item">
            <div className="title">Primary DNS</div>
            <input className="input"
              onChange={e => UpdateInput("DNS", "DNS1", e.target.value)}
              value={inputs["DNS"] ? inputs["DNS"].DNS1 : ""} />
          </div>

          <div className="item item-extra-margin">
            <div className="title">Backup DNS</div>
            <input className="input"
              onChange={e => UpdateInput("DNS", "DNS2", e.target.value)}
              value={inputs["DNS"] ? inputs["DNS"].DNS2 : ""} />
          </div>

          <div className="item extra-space">

            <div onClick={() => UpdateRouterFile(false)} className="title neutral-color">Select A Router File</div>

          </div>

          <div className="item">
            <div className="title neutral-color" onClick={() => UpdateRouterFile(true)} >Clear Router File</div>
          </div>

        </div>

      }


      <div className="panel block-panel">
        <div className="header">
          <LockClosedIcon className="icon"></LockClosedIcon>
          <div className="title">Domain Blocking</div>
          <div className="save-config title neutral-color" onClick={() => ApplyBlocklistConfigurations()}>Save</div>
        </div>


        <div className="item">
          <div className="subtitle">Enabling blocklists will increase memory usage</div>
        </div>

        <div className="item extra-space blocking-toggles">
          <div className="title neutral-color" onClick={() => ToggleAllBlocking(true)} >Enable All</div>
          <div className="title warning-color" onClick={() => ToggleAllBlocking(false)} >Disable All</div>
        </div>
        <div className="item extra-space">
        </div>

        {blockList?.map((item) => {
          return (

            <div className="item less-space">
              <div className="am toggle-button">
                <label className="switch">
                  <input checked={item.Enabled} type="checkbox" onChange={() => ToggleBlockList(item.Tag)} />
                  <span className="slider"></span>
                </label>
                <div className="text">
                  {item.Tag} <span className="subtext">  ( {item.Domains}{' Domains'} )</span>
                </div>
              </div>
            </div>
          )


        })}


      </div>


      {/* <div className="panel capture-panel">

        <div className="header">
          <LockClosedIcon className="icon"></LockClosedIcon>
          <div className="title">Prental Controls</div>
        </div>

        <div className="item">
          {!whitelist &&
            <div className="title  neutral-color" onClick={() => EnableWhitelist()} >Enable Domain Blocking</div>
          }
          {whitelist &&
            <div className="title  neutral-color" onClick={() => DisableWhitelist()} >Disable Domain Blocking</div>
          }
        </div>

        {(props.state?.C && props.state.C?.DomainWhitelist !== "") &&
          <div className="item extra-space">
            <div className="title">Allowed Domains List</div>
            <div className="value">
              {props.state?.C?.DomainWhitelist}
            </div>
          </div>
        }

        <div className="item extra-space">
          {!capture &&
            <div className="title  neutral-color" onClick={() => StartCapture()} >Start Capturing Allowed Domains</div>
          }
          {capture &&
            <div className="title  neutral-color" onClick={() => StopCapture()} >Stop Capturing</div>
          }
        </div>

      </div> */}

      <div className="panel other-panel">

        <div className="header">
          <Pencil2Icon className="icon"></Pencil2Icon>
          <div className="title">Other</div>
          {user &&
            <div className="save-config title neutral-color" onClick={() => UpdateUser()}>Save</div>
          }
        </div>

        <div className="seperator"></div>

        {user &&
          <div className="item">
            <div className="title">Account ID</div>
            <div className="value">{user._id}</div>
          </div>
        }

        {user &&
          <div className="item">
            <div className="title">Cash Payment Code</div>
            <div className="value">{user.CashCode}</div>
          </div>
        }

        {(user && props.advancedMode) &&
          <>
            {inputs["APIKey"] &&
              <div className="item">
                <div className="title">API Key</div>
                <div className="value">{inputs["APIKey"].APIKey}</div>
              </div>
            }
            <div className="item ">
              <div className="title  neutral-color" onClick={() => CreateApiKey()} >Refresh API Key</div>
            </div>
          </>
        }

        {props.state?.LogPath &&
          <div className="item  extra-space">
            <div className="title">Log File</div>
            <div className="value">
              {props.state.LogFileName}
            </div>
          </div>
        }

        {props.state?.ConfigPath &&
          <div className="item">
            <div className="title">Config Backup</div>
            <div className="value">
              {props.state.ConfigPath}
            </div>
          </div>
        }
        {props.state?.C?.RouterFilePath !== "" &&
          <div className="item">
            <div className="title">Router File</div>
            <div className="value">
              {props.state?.C?.RouterFilePath}
            </div>
          </div>
        }

      </div>

      {props.advancedMode &&

        <div className="panel other-panel">

          <div className="header">
            <Pencil2Icon className="icon"></Pencil2Icon>
            <div className="title">Experimental Features</div>
            {user &&
              <div className="save-config title neutral-color" onClick={() => UpdateConfig()}>Save</div>
            }
          </div>

          <div className="item">
            <div className="subtitle">All features in this section are experimental, use with caution</div>
          </div>

          <div className="item extra-space">
            <div className="am toggle-button">
              <label className="switch">
                <input checked={(inputs["CC"] && inputs["CC"].CC)} type="checkbox" onChange={() => ToggleCC()} />
                <span className="slider"></span>
              </label>
              <div className="text">
                Close Sockets On Connect
              </div>
            </div>
            <div className="item">
              <div className="subtitle">This feature will attempt to close TCP sockets when the VPN connects, preventing network leakage outside the VPN connection. Currently this feature only works on Windows</div>
            </div>
          </div>


        </div>
      }




    </div >
  )

}

export default Settings;