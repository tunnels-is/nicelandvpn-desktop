import { HashRouter, Route, Routes } from "react-router-dom";
import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";

import toast, { Toaster } from 'react-hot-toast';
import dayjs from "dayjs";

import "./assets/style/app.scss";

import { CloseApp, IsProduction } from '../wailsjs/go/main/App';
import { Disconnect, LoadRoutersUnAuthenticated, GetRoutersAndAccessPoints, GetState } from '../wailsjs/go/main/Service';

import ScreenLoader from "./App/ScreenLoader";
import DeviceLogins from "./App/DeviceLogins";
import Register from "./App/Registration";
import Dashboard from "./App/Dashboard";
import Enable2FA from "./App/Enable2FA";
import Settings from "./App/Settings";
import Routers from "./App/Routers";
import Support from "./App/Support";
import SideBar from "./App/SideBar";
import Debug from "./App/debug";
import Login from "./App/Login";
import Logs from "./App/Logs";
import STORE from "./store";
import StatsSideBar from "./App/StatsSideBar";

const root = createRoot(document.getElementById('app'));

// window.addEventListener('focus',
//   STORE.Cache.Set("focus", true)
// );

// window.addEventListener('blur',
//   STORE.Cache.Set("focus", false)
// );

const ToggleError = (e) => {
	let lastFetch = STORE.Cache.Get("error-timeout")
	let now = dayjs().unix()
	if ((now - lastFetch) < 3) {
		return
	}
	toast.error(e);
	STORE.Cache.Set("error-timeout", dayjs().unix())
}

const ShowSuccessToast = (e) => {
	toast.success(e);
}

let ShowStartupLoadingScreen = true
let StatupLoadingScreenStartTime = dayjs()

const LaunchApp = () => {
	const [advancedMode, setAdvancedMode] = useState();
	const [loading, setLoading] = useState(undefined)
	const [state, setState] = useState({})
	const [stats, setStats] = useState(false)


	const ToggleAdvancedMode = () => {
		if (STORE.Config.AdvancedMode === true) {
			STORE.Config.AdvancedMode = false
		} else {
			STORE.Config.AdvancedMode = true
		}

		STORE.Cache.Set("advanced", STORE.Config.AdvancedMode)
		setAdvancedMode(STORE.Config.AdvancedMode)
	}

	const ToggleLoading = (object) => {
		if (object?.show) {
			setLoading(object)
		} else {
			const to = setTimeout(() => {
				setLoading(undefined)
			}, 4000)
			return () => { clearTimeout(to) }
		}
	}

	const DisconnectFromVPN = async () => {

		ToggleLoading({ logTag: "disconnect", tag: "LOGOUT", show: true, msg: "Disconnecting", includeLogs: true })

		await Disconnect().then(() => {
			ShowSuccessToast("Disconnected", { Title: "DISCONNECTED", Body: "You have been disconnected from your VPN", TimeoutType: "default" })
			STORE.CleanupOnDisconnect()
		}).catch((e) => {
			console.dir(e)
			ToggleError("Unknown error, please try again in a moment")
		})

		setTimeout(() => {
			setLoading(undefined)
		}, 1000)

	}

	const GetStateAndUpdateVPNList = async () => {
		let newState = { ...state }

		try {

			console.dir(state.ActiveRouter)
			console.log("getting access points")
			if (STORE.ActiveRouterSet(state)) {
				let user = STORE.GetUser()
				if (user) {
					let FR = {
						Method: "POST",
						Path: "devices/private",
						JSONData: {
							UID: user._id,
							DeviceToken: user.DeviceToken.DT
						},
					}
					GetRoutersAndAccessPoints(FR).then((x) => {
						if (x.Code === 401) {
							ToggleError(ERROR_LOGIN)
							STORE.Cache.Clear()
						}

						if (x.Err) {
							ToggleError(x.Err)
						} else {
							if (x.Code !== 200) {
								ToggleError(x.Data)
							}
						}

					}).catch((e) => {
						console.dir(e)
						ToggleError("Unknown error while trying to get VPN list")
					})

				} else {

					LoadRoutersUnAuthenticated().then((x) => {
					}).catch((e) => {
						console.dir(e)
						ToggleError("Unknown error while loading routers un-authenticated")
					})

				}
			}
		} catch (error) {
			console.dir(error)
		}

		try {

			console.log("GET STATE!!!!")
			GetState().then((x) => {
				console.dir(x)
				if (x.Err) {
					ToggleError(x.Err)
					setState(newState)
					return
				}

				if (x.Data) {
					newState = { ...x.Data }
					STORE.Cache.SetObject("state", newState)

					if (newState.C) {
						STORE.Cache.SetObject("config", newState.C)
					}
				}

				setState(newState)

			}).catch(error => {
				console.dir(error)
				ToggleError("Unknown error, please try again in a moment")
				setState(newState)
			});

		} catch (error) {
			console.dir(error)
			setState(newState)
		}

	}

	const UpdateAdvancedMode = () => {
		let status = STORE.AdvancedModeEnabled()
		if (status !== advancedMode) {
			console.log("UPDATING ADVANCED MODE:", status, advancedMode)
			setAdvancedMode(status)
		}
	}

	useEffect(() => {

		if (ShowStartupLoadingScreen) {
			setLoading({ logTag: "loader", tag: "READY", show: true, msg: "Starting Niceland", includeLogs: true })

			let now = dayjs()
			if (now.diff(StatupLoadingScreenStartTime, "s") > 10) {
				ShowStartupLoadingScreen = false
				setLoading(undefined)
			}

			if (state && state.ClientReady) {
				if (loading && loading.tag === "READY") {
					ShowStartupLoadingScreen = false
					setLoading(undefined)
				}
			}
		}

		const to = setTimeout(async () => {
			UpdateAdvancedMode()
			GetStateAndUpdateVPNList()
		}, 1000)

		return () => { clearTimeout(to); }

	}, [state]);

	return (
		< HashRouter >
			<>

				<Toaster
					containerStyle={{
						top: "20px",
						left: "20px",
						position: 'fixed',
					}}
					toastOptions={{
						className: 'toast',
						position: "top-left",
						success: {
							duration: 5000,
						},
						icon: null,
						error: {
							duration: 5000,
							style: {
							},
						},
					}}
				/>

				{loading &&
					<ScreenLoader loading={loading} toggleError={ToggleError}></ScreenLoader>
				}

				{/* <TopBar toggleLoading={ToggleLoading}></TopBar> */}
				<SideBar advancedMode={advancedMode} toggleLoading={ToggleLoading} state={state} loading={loading} disconnectFromVPN={DisconnectFromVPN} toggleError={ToggleError} setStats={setStats} stats={stats} />

				{stats &&
					<StatsSideBar state={state} setStats={setStats}></StatsSideBar>
				}


				<div className="content-container" >

					<Routes>

						<Route path="/" element={<Dashboard state={state} advancedMode={advancedMode} toggleLoading={ToggleLoading} toggleError={ToggleError} showSuccessToast={ShowSuccessToast} disconnectFromVPN={DisconnectFromVPN} />} />
						<Route path="twofactor" element={<Enable2FA toggleError={ToggleError} toggleLoading={ToggleLoading} />} />
						<Route path="support" element={<Support toggleError={ToggleError} />} />
						<Route path="settings" element={<Settings advancedMode={advancedMode} showSuccessToast={ShowSuccessToast} toggleAdvancedMode={ToggleAdvancedMode} toggleError={ToggleError} disconnectFromVPN={DisconnectFromVPN} toggleLoading={ToggleLoading} state={state} />} />
						<Route path="tokens" element={<DeviceLogins toggleError={ToggleError} showSuccessToast={ShowSuccessToast} toggleLoading={ToggleLoading} />} />
						<Route path="logs" element={<Logs toggleError={ToggleError} />} />
						<Route path="debug" element={<Debug toggleError={ToggleError} showSuccessToast={ShowSuccessToast} toggleLoading={ToggleLoading} />} />
						<Route path="login" element={<Login state={state} toggleError={ToggleError} showSuccessToast={ShowSuccessToast} toggleLoading={ToggleLoading} />} />
						<Route path="register" element={<Register toggleError={ToggleError} showSuccessToast={ShowSuccessToast} />} />
						<Route path="routers" element={<Routers state={state} toggleLoading={ToggleLoading} toggleError={ToggleError} showSuccessToast={ShowSuccessToast} />} />
						<Route path="*" element={<Dashboard state={state} advancedMode={advancedMode} toggleLoading={ToggleLoading} toggleError={ToggleError} showSuccessToast={ShowSuccessToast} disconnectFromVPN={DisconnectFromVPN} />} />

					</Routes>
				</div>
			</>

		</HashRouter >
	)


}




class ErrorBoundary extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			hasError: false,
			title: "Something unexpected happened, please press Reload. If that doesn't work try pressing 'Close And Reset'. If nothing works, please contact customer support"
		};
	}

	static getDerivedStateFromError() {
		return { hasError: true };
	}

	componentDidCatch() {
		this.state.hasError = true
	}

	reloadAll() {
		// STORE.Cache.Clear()
		window.location.reload()
	}

	async quit() {
		this.setState({ ...this.state, title: "closing app, please wait.." })
		await Disconnect().then(() => {
		}).catch((e) => {
			console.dir(e)
		})
		STORE.Cache.Clear()
		CloseApp()
	}

	async ProductionCheck() {

		await IsProduction().then((x) => {
			if (x) {
				var console = {}
				console.apply = function() { }
				console.log = function() { }
				console.dir = function() { }
				console.info = function() { }
				console.warn = function() { }
				console.error = function() { }
				window.console = console
			}

		}).catch((e) => {
			console.dir(e)
		})
	}

	render() {

		this.ProductionCheck()

		if (this.state.hasError) {
			return (<>
				<h1 className="exception-title">
					{this.state.title}
				</h1>
				<button className="exception-button" onClick={() => this.reloadAll()}>Reload</button>
				<button className="exception2-button" onClick={() => this.quit()}>Close And Reset</button>
			</>)
		}

		return this.props.children;
	}
}

STORE.Cache.Set("focus", true)

root.render(
	<React.StrictMode>
		<ErrorBoundary>
			<LaunchApp />
		</ErrorBoundary>
	</React.StrictMode>
)
