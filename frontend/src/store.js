import dayjs from "dayjs";
const DATA = "data_";

var STORE = {
  Config: {
    EnableConsoleLogs: false,
    AdvancedMode: false,
  },
  ROUTER_Tooltips: [
    "Quality of service is a score calculated from latency, available bandwidth and number of available user spots on the Router. 10 is the best score, 0 is the worst score",

    "Latency from your computer to this Router",

    "Available user slots",

    "Available Gigabits per second of bandwidth",

    "Processor usage in %",

    "Memory usage in %",

    "Disk usage in %",

    "",
    "",
  ],
  VPN_Tooltips: [
    "Quality of service is a score calculated from latency, available bandwidth and number of available user spots on the VPN's Router. 10 is the best score, 0 is the worst score",

    "Available user slots on the VPN's Router ( Total / Available )",

    "Available bandwidth in % on the VPN's Router ( Download / Upload )",

    "Processor usage in percentages on the VPN's Router",

    "Memory usage in percentages on the VPN's Router",

    "Available Gigabits per second of bandwidth on the VPN's Router",
  ],
  GetUser() {
    let user = STORE.Cache.GetObject("user")
    if (!user) {
      return undefined
    }
    return user
  },
  CleanupOnDisconnect() {
    // We might want to add some cleanup in here at a later stage
  },
  CleanupOnLogout() {
    let user = STORE.Cache.GetObject("user")
    if (user) {
      STORE.Cache.DelObject(user.Email + "_TOKEN")
    }

    STORE.Cache.DelObject("user");
  },
  CleanupOnClose() {
    STORE.Cache.DelObject("state")
    let AL = STORE.Cache.GetBool("auto-logout")
    if (AL) {
      STORE.CleanupOnLogout()
    }
  },
  ActiveRouterSet(state) {
    if (!state) {
      return false
    } else if (!state.ActiveRouter) {
      return false
    } else if (state.ActiveRouter.PublicIP === "") {
      return false
    }
    return true
  },
  AdvancedModeEnabled: function () {
    let as = STORE.Cache.GetBool("advanced")
    if (as === true) {
      STORE.Config.AdvancedMode = true
      return true
    } else {
      STORE.Config.AdvancedMode = false
    }
    return STORE.Config.AdvancedMode

  },
  Cache: {
    MEMORY: {
      FetchingState: false,
      DashboardData: undefined,
    },
    Clear: function (key) {
      return window.localStorage.clear()
    },
    Get: function (key) {
      return window.localStorage.getItem(key)
    },
    GetBool: function (key) {
      let data = window.localStorage.getItem(key)
      if (data === "true") {
        return true
      }
      return false
    },
    Set: function (key, value) {
      window.localStorage.setItem(key, value)
    },
    Del: function (key) {
      window.localStorage.removeItem(key)
    },
    DelObject: function (key) {
      window.localStorage.removeItem(DATA + key)
      window.localStorage.removeItem(DATA + key + "_ct")
    },
    GetObject: function (key) {
      let jsonData = null
      try {
        let object = window.localStorage.getItem(DATA + key)
        if (object === "undefined") {
          return undefined
        } else {

          jsonData = JSON.parse(object)
          if (STORE.EnableConsoleLogs) {
            console.log("%cGET OBJECT:", 'background: lightgreen; color: black', key, jsonData)
          }
        }
      }
      catch (e) {
        console.log(e)
        return undefined
      }

      if (jsonData === null) {
        return undefined
      }

      return jsonData
    },
    SetObject: function (key, object) {
      try {
        if (STORE.EnableConsoleLogs) {
          console.log("%cSET OBJECT:", 'background: lightgreen; color: black', key, object)
        }
        let data = JSON.stringify(object)
        window.localStorage.setItem(DATA + key, data)
        window.localStorage.setItem(DATA + key + "_ct", dayjs().unix())
      }
      catch (e) {
        console.log(e)
        alert(e)
      }

    },
    GetCatchTimer(key) {
      try {
        return window.localStorage.getItem(DATA + key + "_ct")
      }
      catch (e) {
        console.log(e)
        alert(e)
      }
      return undefined
    }
  },

};


export default STORE;