import { HashRouter, Route, Routes } from "react-router-dom";
import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";

import toast, { Toaster } from 'react-hot-toast';
import dayjs from "dayjs";

import "./assets/style/app.scss";

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
import API from "./api";
import Connections from "./App/Connections";

const root = createRoot(document.getElementById('app'));

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
			return () => {
				clearTimeout(to)
			}
		}
	}

	const DisconnectFromVPN = async () => {

		let x = await API.method("disconnect", undefined)
		if (x === undefined) {
			console.dir(e)
			ToggleError("Unknown error, please try again in a moment")
			setLoading(undefined)
		} else {
			ShowSuccessToast("Disconnected", {
				Title: "DISCONNECTED", Body: "You have been disconnected from your VPN", TimeoutType: "default"
			})
			STORE.CleanupOnDisconnect()
			setLoading(undefined)
		}

	}

	const GetStateAndUpdateVPNList = async () => {
		let newState = { ...state }

		try {
			let user = STORE.GetUser()
			let FR = {}
			if (user) {
				FR = {
					Method: "POST", Path: "devices/private", Authed: true, JSONData: {
						UID: user._id, DeviceToken: user.DeviceToken.DT
					},
				}
			}

			let x = await API.method("getState", FR)
			console.dir(x)
			if (x === undefined) {
				ToggleError("Unknown error, please try again in a moment")
			} else {
				newState = { ...x.data }

				const nodes = newState.Nodes.filter((node) => {
					if (node !== null) {
						return true
					}
					return false
				})
				newState.Nodes = nodes

				const pnodes = newState.PrivateNodes.filter((node) => {
					if (node !== null) {
						return true
					}
					return false
				})
				newState.PrivateNodes = pnodes

				const routers = newState.Routers.filter((node) => {
					if (node !== null) {
						return true
					}
					return false
				})
				newState.Routers = routers


				STORE.Cache.SetObject("state", newState)

				// if (newState.C) {
				// 	STORE.Cache.SetObject("config", newState.C)
				// 	STORE.Cache.SetObject("config", newState.C)
				// }
				setState(newState)

			}

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

		const to = setTimeout(async () => {
			UpdateAdvancedMode()
			await GetStateAndUpdateVPNList()
		}, 2000)

		return () => {
			clearTimeout(to);
		}

	}, [state]);

	return (< HashRouter>
		<>

			<Toaster
				containerStyle={{
					top: "20px", left: "20px", position: 'fixed',
				}}
				toastOptions={{
					className: 'toast', position: "top-right", success: {
						duration: 5000,
					}, icon: null, error: {
						duration: 5000, style: {},
					},
				}}
			/>

			{loading && <ScreenLoader loading={loading} toggleError={ToggleError}></ScreenLoader>}

			{/* <TopBar toggleLoading={ToggleLoading}></TopBar> */}
			<SideBar advancedMode={advancedMode} toggleLoading={ToggleLoading} state={state} loading={loading}
				disconnectFromVPN={DisconnectFromVPN} toggleError={ToggleError} setStats={setStats}
				stats={stats} />

			{stats && <StatsSideBar state={state} setStats={setStats}></StatsSideBar>}


			<div className="content-container">

				<Routes>

					<Route path="/" element={<Dashboard
						state={state}
						advancedMode={advancedMode}
						toggleLoading={ToggleLoading}
						toggleError={ToggleError}
						showSuccessToast={ShowSuccessToast}
						disconnectFromVPN={DisconnectFromVPN} />} />

					<Route path="twofactor" element={<Enable2FA
						toggleError={ToggleError}
						toggleLoading={ToggleLoading} />} />

					<Route path="support" element={<Support
						toggleError={ToggleError} />} />

					<Route path="settings" element={<Settings
						advancedMode={advancedMode}
						showSuccessToast={ShowSuccessToast}
						toggleAdvancedMode={ToggleAdvancedMode}
						toggleError={ToggleError}
						disconnectFromVPN={DisconnectFromVPN}
						toggleLoading={ToggleLoading}
						state={state} />} />

					<Route path="tokens" element={<DeviceLogins
						toggleError={ToggleError}
						showSuccessToast={ShowSuccessToast}
						toggleLoading={ToggleLoading} />} />

					<Route path="logs" element={<Logs
						toggleError={ToggleError} />} />

					<Route path="debug" element={<Debug
						toggleError={ToggleError}
						showSuccessToast={ShowSuccessToast}
						toggleLoading={ToggleLoading} />} />

					<Route path="login" element={<Login
						state={state}
						toggleError={ToggleError}
						showSuccessToast={ShowSuccessToast}
						toggleLoading={ToggleLoading} />} />

					<Route path="register" element={<Register
						toggleError={ToggleError}
						showSuccessToast={ShowSuccessToast} />} />

					<Route path="routers" element={<Routers
						state={state}
						toggleLoading={ToggleLoading}
						toggleError={ToggleError}
						showSuccessToast={ShowSuccessToast} />} />

					<Route path="connections" element={<Connections
						state={state}
						toggleLoading={ToggleLoading}
						toggleError={ToggleError}
						showSuccessToast={ShowSuccessToast} />} />

					<Route path="*" element={<Dashboard
						state={state}
						advancedMode={advancedMode}
						toggleLoading={ToggleLoading}
						toggleError={ToggleError}
						showSuccessToast={ShowSuccessToast}
						disconnectFromVPN={DisconnectFromVPN} />} />

				</Routes>
			</div>
		</>

	</HashRouter>)


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

		// var console = {}
		// console.apply = function() {
		// }
		// console.log = function() {
		// }
		// console.dir = function() {
		// }
		// console.info = function() {
		// }
		// console.warn = function() {
		// }
		// console.error = function() {
		// }
		// console.debug = function() {
		// }
		// window.console = console

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

root.render(<React.StrictMode>
	<ErrorBoundary>
		<LaunchApp />
	</ErrorBoundary>
</React.StrictMode>)
