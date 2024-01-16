import { Navigate } from "react-router-dom";
import React, { useEffect, useState } from "react";


import Loader from "react-spinners/ScaleLoader";
import STORE from "../store";
import { CheckIcon, ChevronDownIcon, ChevronUpIcon, ExitIcon, EyeOpenIcon, FileTextIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import API from "../api";

import Editor, { useMonaco } from '@monaco-editor/react';
import { loader } from '@monaco-editor/react';

import * as monaco from 'monaco-editor';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';


const Connections = (props) => {
	loader.config({ monaco });
	self.MonacoEnvironment = {
		getWorker(_, label) {
			return new jsonWorker();
		},
	};


	const [filter, setFilter] = useState("");
	const [connections, setConnections] = useState([])
	const [editor, setEditor] = useState(undefined)
	const [changed, setChanged] = useState(false)
	const [openEditors, setOpenEditors] = useState(new Map())

	const openEditor = (id) => {
		let e = openEditors
		e.set(id, true)
		setOpenEditors(new Map(e))
	}

	const closeEditor = (id) => {
		let e = openEditors
		e.set(id, false)
		setOpenEditors(new Map(e))
		setChanged(false)
	}

	const renderRouterTooltip = (r) => {
		return (
			<div className="router-tooltip">
				<div className="tag">{r.Tag}</div>
				<div className="ip">{r.IP}</div>
				<div className="ms">MS: {r.MS}</div>
				<div className="ram">Country Code: {r.Country}</div>
				<div className="ram">Mbps: {r.AvailableMbps}</div>
				<div className="disk">DISK: {r.DiskUsage}</div>
				<div className="cpu">CPU: {r.CPUP}</div>
				<div className="ram">RAM: {r.RAMUsage}</div>
				<div className="ram">Slots: {r.Slots}</div>
				<div className="ram">Index: {r.ListIndex}</div>

			</div>
		)
	}

	const renderNodeTooltip = (n) => {
		return (
			<div className="router-tooltip">
				<div className="tag">{n.Tag}</div>
				<div className="ip">{n.IP}</div>
				<div className="id">ID: {n._id}</div>
				<div className="ms">MS: {n.MS}</div>
				<div className="ram">Country Code: {n.Country}</div>
				<div className="ram">Mbps: {n.AvailableMbps}</div>
				<div className="ram">Slots: {n.Slots}</div>
				<div className="ram">Online: {n.Online}</div>

			</div>
		)
	}

	const save = async (id) => {

		let connections = []
		let editors = editor.editor.getEditors()
		let er = false
		editors.forEach((e) => {
			console.log("SINGLE EDITOR!")
			try {
				connections.push(JSON.parse(e.getValue()))
			} catch (error) {
				props.toggleError("cannot save config: invalid json")
				er = true
			}
		})
		if (er === true) {
			return
		}
		// console.log("POST UPDATE")
		// console.dir(props.state?.C)
		// console.dir(props.state?.C?.Connections)
		// console.dir(connections)
		// return

		let newConfig = { ...props.state?.C }
		newConfig.Connections = connections

		let resp = await API.method("setConfig", newConfig)
		if (resp === undefined) {
			props.toggleError("Unknown error, please try again in a moment")
		} else {
			if (resp.status === 200) {
				closeEditor(id)
				props.showSuccessToast("Config saved")
			} else {
				console.dir(resp.data)
				props.toggleError(resp.data)
			}
		}
		setConnections(connections)
		setChanged(false)
	}

	const setEditorTheme = (monaco) => {
		monaco.editor.defineTheme('onedark', {
			base: 'vs-dark',
			inherit: true,
			// colors: {
			// 	'editor.background': '#000000'
			// }
		});
	}


	const monacoMount = (editor) => {
		console.log("MOUNT")
		console.dir(editor)
		editor.onKeyDown((event) => {
			console.log("event", event)
			setChanged(true)
		})
	}

	useEffect(() => {
		loader.init().then((e) => {
			console.log("MONOCO INIT")
			// let editors = e.editor.getEditors()
			// console.dir(editors)
			if (editor === undefined) {
				setEditor(e)
			}
		})
		setConnections(props.state?.C?.Connections)
	}, [props.state?.C?.Connections])


	const LogOut = () => {
		props.toggleError("You have been logged out")
		STORE.Cache.Clear()
	}

	const ConnectToVPN = async (connection) => {

		try {

			props.toggleLoading({ logTag: "connect", tag: "CONNECT", show: true, msg: "Initializing " + connection.Tag, includeLogs: true })

			let user = STORE.GetUser()
			if (!user) {
				LogOut()
				return (<Navigate to={"/login"} />)
			}

			// let method = undefined
			// if (props.state?.ActiveAccessPoint) {
			// 	method = "switch"
			// } else {
			let method = "connect"
			// }


			let connectionRequest = {}
			connectionRequest.ID = connection._id
			connectionRequest.UserID = user._id
			connectionRequest.DeviceToken = user.DeviceToken.DT


			let resp = await API.method(method, connectionRequest)
			if (resp === undefined) {
				props.toggleError("Unknown error, please try again in a moment")
			} else {
				if (resp.status === 401) {
					LogOut()
				} else if (resp.status === 200) {
					props.showSuccessToast(connection.Tag + " initialized", undefined)
				} else if (resp.data) {
					console.log("HAHSJDHAKSJHDKHASDKJ")
					console.dir(resp.data)
					props.toggleError(resp.data)
				} else {
					props.toggleError("Unknown error, please try again in a moment")
				}
			}

		} catch (error) {
			console.dir(error)
		}

		props.toggleLoading(undefined)
	}


	const updateRoute = (c, type, key, newValue) => {

		// console.dir(id)
		console.dir(c)
		console.dir(newValue)

		let consx = [...connections]
		consx.forEach((con) => {
			if (con.Name !== c.Name) {
				return
			}

			con.Routes.forEach((route) => {

				if (type === "route") {
					if (route.Name === key) {
						route.Route = newValue
					}
				} else {
					if (route.Route === key) {
						route.Name = newValue
					}
				}

			})
		})

		STORE.Cache.SetObject("connections", consx)
		setConnections([...consx])

	}

	const toggle = (c, id) => {
		let consx = [...connections]
		consx.forEach((con) => {
			if (con.Name === c.Name) {
				con[id] = !con[id]
			}
		})
		STORE.Cache.SetObject("connections", consx)
		setConnections([...consx])
	}

	const inputChange = (c, id, newValue) => {
		let consx = [...connections]
		consx.forEach((con) => {
			if (con.Name === c.Name) {
				con[id] = newValue
			}
		})
		STORE.Cache.SetObject("connections", consx)
		setConnections([...consx])
	}

	const RenderConnection = (c) => {
		return (
			<div className="connection" key={c.Tag} >
				<div className="name">
					{c.Tag}
				</div>
				<div className="small-title">
					{c._id}
				</div>
				<div className="connect" onClick={() => ConnectToVPN(c)}>
					CONNECT
				</div>

				<div className="title">Tunnel</div>

				<div className="cell">Interface
					<input
						className="value"
						id="IFName"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.IFName}
					/>
				</div>

				<div className="cell">IPv4
					<input
						className="value"
						id="IPv4Address"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.IPv4Address}
					/>
				</div>
				<div className="cell">NetworkMask
					<input
						className="value"
						id="NetMask"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.NetMask}
					/>
				</div>

				<div className="cell">MTU
					<input
						className="value"
						id="MTU"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.MTU}
					/>
				</div>

				<div className="cell">TXQueueLen
					<input
						className="value"
						id="TxQueueLen"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.TxQueueLen}
					/>
				</div>

				<div className="cell">Encryption
					<input
						className="value"
						id="EncryptionProtocol"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.EncryptionProtocol}
					/>
				</div>

				<div className="cell">AutoReconnect
					<div
						className={`value toggle special-fix ${c.AutoReconnect ? "on" : "off"}`}
						id="AutoReconnect"
						onClick={(e) => toggle(c, e.target.id)}
					>{c.AutoReconnect ? "enabled" : "disabled"}</div>
				</div>
				<div className="cell">AutoRouter
					<div
						className={`value toggle special-fix ${c.AutomaticRouter ? "on" : "off"}`}
						id="AutomaticRouter"
						onClick={(e) => toggle(c, e.target.id)}
					>{c.AutomaticRouter ? "enabled" : "disabled"}</div>
				</div>

				<div className="cell">Persistent
					<div
						className={`value toggle ${c.Persistent ? "on" : "off"}`}
						id="Persistent"
						onClick={(e) => toggle(c, e.target.id)}
					>
						{c.Persistent ? "yes" : "no"}
					</div>
				</div>

				<div className="cell">Killswitch
					<div
						className={`value toggle ${c.Killswitch ? "on" : "off"}`}
						id="Killswitch"
						onClick={(e) => toggle(c, e.target.id)}
					>
						{c.Killswitch ? "enabled" : "disabled"}
					</div>
				</div>



				<div className="cell">Country
					<select name="Country" id="Country" value={c.Country}>
						<option defaultChecked className="hidden" key={"none"} value={"none"}>Select Country</option>
						{props.state?.AvailableCountries?.map((r) => {
							if (r === c.Country) {
								return (
									<option className="hidden" key={r} value={r}>{r}</option>
								)
							} else {
								return (
									<option key={r} value={r}>{r}</option>
								)
							}
						})}
					</select>
				</div>

				<div className="title">DNS Settings</div>
				<div className="cell">EnableDNS
					<div
						className={`value toggle ${c.CustomDNS ? "on" : "off"}`}
						id="CustomDNS"
						onClick={(e) => toggle(c, e.target.id)}
					>
						{c.CustomDNS ? "enabled" : "disabled"}
					</div>
				</div>
				<div className="cell">DNS1
					<input
						className="value"
						id="DNS1"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.DNS1}
					/>
				</div>
				<div className="cell">DNS2
					<input
						className="value"
						id="DNS2"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.DNS2}
					/>
				</div>

				<div className="title">Networks</div>
				{c.Networks?.map((n) => {
					return (
						<div key={n._id}>
							<div className="cell" >Tag
								<input
									className="value"
									id="nat"
									// onChange={(e) => inputChange(c, e.target.id, e.target.value)}
									disabled
									value={n.Tag}
								/>
							</div>

							<div className="cell" >Network
								<input
									className="value"
									id="Network"
									// onChange={(e) => inputChange(c, e.target.id, e.target.value)}
									disabled
									value={n.Network}
								/>
							</div>

							<div className="cell" >Nat
								<input
									className="value"
									id="nat"
									// onChange={(e) => inputChange(c, e.target.id, e.target.value)}
									value={n.Nat} />
							</div>


							<div className="sub-title">Routes</div>
							{
								n.Routes?.map((r) => {
									return (
										<>
											<div className="cell" >Address
												<input
													className="value"
													id="Address"
													onChange={(e) => updateRoute(c, "Address", r.Route, e.target.value)}
													value={r.Address}
												/>
											</div>
											<div className="cell" >Metric
												<input
													id="Metric"
													className="value"
													onChange={(e) => updateRoute(c, "Metric", r.Name, e.target.value)}
													value={r.Metric}
												/>
											</div>
										</>
									)
								})
							}


						</div>
					)
				})
				}


				<div className="title">( Entry ) Router</div>
				<div className="cell">Tag
					<select name="Router" id="Router" value={c.RouterIndex}>
						<option defaultChecked className="hidden" key={"none"} value={"none"}>Select Router</option>
						{props.state?.Routers?.map((r) => {
							if (r.ListIndex === c.RouterIndex) {
								return (
									<option className="hidden" key={r.Tag} value={r.ListIndex}>{r.Tag}</option>
								)
							} else {
								return (
									<option key={r.Tag} value={r.ListIndex}>{r.Tag}</option>
								)
							}
						})}
					</select>
				</div>

				{
					props.state?.Routers?.map((x) => {
						if (x.ListIndex === c.RouterIndex) {
							return (
								<>
									<div className="cell">IP
										<input className="value" disabled value={x.PublicIP} />
									</div>
									<div className="cell">Mbps
										<input className="value" disabled value={x.AvailableMbps} />
									</div>
									<div className="cell">MS
										<input className="value" disabled value={x.MS} />
									</div>
									<div className="cell">Availability(slots)
										<input className="value" disabled value={x.AvailableSlots} />
									</div>
								</>
							)
						}
					})
				}

				<div className="title">( Exit ) Node</div>
				<div className="cell">Tag

					<select name="Node" id="Node" value={c.NodeID} >
						<option defaultChecked className="hidden" key={"none"} value={"none"}>Select Node</option>
						{props.state?.PrivateNodes?.map((r) => {
							if (r._id === c.NodeID) {
								return (
									<option className="hidden" key={r.ListIndex} value={r.ListIndex}>{r.Tag} </option>
								)
							} else {
								return (
									<option key={r.ListIndex} value={r.ListIndex}>{r.Tag}</option>
								)
							}
						})}
						{props.state?.PrivateNodes?.length > 0 &&
							<option disabled key={"seperator"} value={"------"}>----------</option>
						}
						{props.state?.Nodes?.map((r) => {
							if (r._id === c.NodeID) {
								return (
									<option className="hidden" key={r.ListIndex} value={r.ListIndex}>{r.Tag}</option>
								)
							} else {
								return (
									<option key={r.ListIndex} value={r.ListIndex}>{r.Tag}</option>
								)
							}
						})}
					</select>
				</div>

				{
					props.state?.Nodes?.map((x) => {
						if (x._id === c.NodeID) {
							return (
								<>
									<div className="cell">IP
										<input className="value" disabled value={x.IP} />
									</div>
									<div className="cell">Mbps
										<input className="value" disabled value={x.AvailableMbps} />
									</div>
									<div className="cell">MS
										<input className="value" disabled value={x.MS} />
									</div>
									<div className="cell">Availability(slots)
										<input className="value" disabled value={x.AvailableSlots} />
									</div>
								</>
							)
						}
					})
				}
				{
					props.state?.PrivateNodes?.map((x) => {
						if (x._id === c.NodeID) {
							return (
								<>
									<div className="cell">IP
										<input className="value" disabled value={x.IP} />
									</div>
									<div className="cell">Mbps
										<input className="value" disabled value={x.AvailableMbps} />
									</div>
									<div className="cell">MS
										<input className="value" disabled value={x.MS} />
									</div>
									<div className="cell">Availability(slots)
										<input className="value" disabled value={x.AvailableSlots} />
									</div>
								</>
							)
						}
					})
				}

			</div >
		)
	}

	const renderConnection = (c) => {

		let entryRouter = undefined
		let proxyRouter = undefined
		let node = undefined
		props.state?.Routers?.map((x) => {
			if (x.ListIndex === c.RouterIndex) {
				entryRouter = x
			}
		})

		props.state?.Nodes?.map((x) => {
			if (x._id === c.NodeID) {
				node = x
			}
		})
		props.state?.PrivateNodes?.map((x) => {
			if (x._id === c.NodeID) {
				node = x
			}
		})

		if (node) {
			props.state?.Routers?.map((x) => {
				if (x.ListIndex === node.RouterIndex) {
					proxyRouter = x
				}
			})
		}
		return (
			<div className={`${openEditors.get(c._id) ? "editor-show" : "editor-hide"} info-flex `}
			>
				<div className="info">
					{openEditors.get(c._id) &&
						<div className="close button" onClick={() =>
							closeEditor(c._id)}>
							<ChevronUpIcon height={30} width={30}></ChevronUpIcon>
						</div>
					}
					{!openEditors.get(c._id) &&
						<div className="edit button" onClick={() =>
							openEditor(c._id)}>
							<ChevronDownIcon height={30} width={30}></ChevronDownIcon>
						</div>
					}

					{c.Tag &&
						<div
							className="tag"
							onClick={() => ConnectToVPN(c)}>
							{c.Tag}
						</div>
					}

					{c.AutomaticRouter &&
						<div className="auto-router">
							{props?.state?.ActiveRouter?.Tag}
							<span className="tooltiptext">
								{renderRouterTooltip(props.state.ActiveRouter)}
							</span>
						</div>
					}

					{(entryRouter && !c.AutomaticRouter) &&
						<div className="ri">
							{entryRouter?.Tag}
							<span className="tooltiptext">
								{renderRouterTooltip(entryRouter)}
							</span>
						</div>
					}

					{proxyRouter &&
						<div className="pr">
							{proxyRouter.Tag}
							<span className="tooltiptext">
								{renderRouterTooltip(proxyRouter)}
							</span>
						</div>
					}
					{node &&
						<div className="tag">
							{node?.Tag}
							<span className="tooltiptext">
								{renderNodeTooltip(node)}
							</span>
						</div>
					}

					{(changed && openEditors.get(c._id)) &&
						<div className="save button" onClick={() =>
							save(c._id)}>
							<CheckIcon height={30} width={30}></CheckIcon>
						</div>
					}
				</div>

				<Editor
					height="90vh"
					options={{
						automaticLayout: true,
						glyphMargin: false,
						roundedSelection: true,
						folding: false,
						lineHeight: 18,
						scrollBeyondLastLine: false,
						lineDecorationsWidth: 2,
						lineNumbersMinChars: 3,
						minimap: {
							enabled: false
						}
					}}
					className="monoco-editor"
					defaultLanguage="json"
					theme="vs-dark"
					onMount={(e) => monacoMount(e)}
					defaultValue={"// Create a connection here"}
					value={JSON.stringify(c, null, 4)}
				/>

			</div >

		)

	}


	return (
		<div className="connection-wrapper"  >
			<div className="connections-flex">

				{(!connections || connections.length < 1) &&
					<Loader
						className="spinner"
						loading={true}
						color={"#20C997"}
						height={100}
						width={50}
					/>
				}

				{connections?.map((c) => {
					return renderConnection(c)
				})}

			</div>
		</div >
	);
}

export default Connections;
