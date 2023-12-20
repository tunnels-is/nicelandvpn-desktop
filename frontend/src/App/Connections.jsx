import { Navigate } from "react-router-dom";
import React, { useState } from "react";


import STORE from "../store";
import { DesktopIcon, MagnifyingGlassIcon, EnterIcon } from "@radix-ui/react-icons";


var xxxxx = [
	{
		IPv4Address: "10.4.3.2",
		IPv6Address: "",
		Name: "Niceland",
		MTU: 1500,
		TxQueueLen: 3000,
		Persistent: true,
		Routes: [
			{ Name: "default", Route: "0.0.0.0/0" },
		],
		Router: "london-01",
		Node: "finland-01",
		AutoReconnect: false,
		DNS: ["1.1.1.1", "9.9.9.9"],
	},
	{
		IPv4Address: "10.4.3.3",
		IPv6Address: "",
		Name: "cloud-edge-router",
		MTU: 1500,
		TxQueueLen: 3000,
		Persistent: true,
		Routes: [
			{ Name: "app-network", Route: "172.17.1.0/24" },
			{ Name: "db-network", Route: "172.18.1.0/24" }
		],
		Router: "finland-01",
		Node: "edge-router",
		AutoReconnect: true,
		DNS: ["1.1.1.1", "9.9.9.9"],
	}
]

const Connections = (props) => {
	STORE.Cache.SetObject("connections", xxxxx)

	const [filter, setFilter] = useState("");
	const [connections, setConnections] = useState(STORE.Cache.GetObject("connections"))

	let cons = []

	if (connections) {

		if (filter && filter !== "") {
			connections.map(r => {

				let filterMatch = false
				Object.keys(r).forEach((key, item) => {
					if (r[key].toString().includes(filter)) {
						filterMatch = true
					}
				})

				if (filterMatch) {
					cons.push(r)
				}

			})

		} else {
			cons = connections
		}

	}

	const updateDNS = (c, index, newValue) => {

		console.dir(index)
		console.dir(c)
		console.dir(newValue)

		let consx = [...connections]
		consx.forEach((con) => {
			if (con.Name !== c.Name) {
				return
			}

			con.DNS[index] = newValue
		})

		STORE.Cache.SetObject("connections", consx)
		setConnections([...consx])

	}


	const updateRoute = (c, type, key, newValue) => {

		console.dir(id)
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
						route.Name = key
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
			<div className="connection" key={c.Name} onClick={() => console.dir("clicked")}>
				<div className="name">
					{c.Name}
				</div>
				<div className="connect">CONNECT</div>

				<div className="title">Tunnel</div>
				<div className="cell">IPv4
					<input
						className="value"
						id="IPv4Address"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.IPv4Address}
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

				<div className="cell">AutoReconnect
					<div
						className={`value toggle special-fix ${c.AutoReconnect ? "on" : "off"}`}
						id="AutoReconnect"
						onClick={(e) => toggle(c, e.target.id)}
					>{c.AutoReconnect ? "enabled" : "disabled"}</div>
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

				<div className="title">Routes</div>
				{
					c.Routes.map((r) => {
						return (
							<div className="cell" key={r.Name}>
								<input
									className="value route-name"
									onChange={(e) => updateRoute(c, "name", r.Route, e.target.value)}
									value={r.Name}
								/>
								<input
									className="value route"
									onChange={(e) => updateRoute(c, "route", r.Name, e.target.value)}
									value={r.Route}
								/>
							</div>
						)
					})
				}

				<div className="title">DNS</div>
				<div className="cell">
					<input
						className="value dns0"
						onChange={(e) => updateDNS(c, 0, e.target.value)}
						value={c.DNS[0]}
					/>
					<input
						className="value dns1"
						onChange={(e) => updateDNS(c, 1, e.target.value)}
						value={c.DNS[1]}
					/>
				</div>

				<div className="title">( Entry ) Router</div>
				<div className="cell">Tag
					<select name="Router" id="Router" value={c.Router}>
						<option defaultChecked className="hidden" key={"none"} value={"none"}>Select Router</option>
						{props.state?.Routers?.map((r) => {
							if (r.Tag === c.Router) {
								return (
									<option className="hidden" key={r.Tag} value={r.Tag}>{r.Tag}</option>
								)
							} else {
								return (
									<option key={r.Tag} value={r.Tag}>{r.Tag}</option>
								)
							}
						})}
					</select>
				</div>

				{props.state?.Routers?.map((x) => {
					if (x.Tag === c.Router) {
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
				})}

				<div className="title">( Exit ) Node</div>
				<div className="cell">Tag
					<select name="Node" id="Node" value={c.Node} >
						<option defaultChecked className="hidden" key={"none"} value={"none"}>Select Node</option>
						{props.state?.AccessPoints?.map((r) => {
							if (r.Tag === c.Node) {
								return (
									<option className="hidden" key={r.Tag} value={r.Tag}>{r.Tag}</option>
								)
							} else {
								return (
									<option key={r.Tag} value={r.Tag}>{r.Tag}</option>
								)
							}
						})}
					</select>
				</div>

				{props.state?.AccessPoints?.map((x) => {
					if (x.Tag === c.Node) {
						return (
							<>
								<div className="cell">IP
									<input className="value" disabled value={x.Router.PublicIP} />
								</div>
								<div className="cell">Mbps
									<input className="value" disabled value={x.Router.AvailableMbps} />
								</div>
								<div className="cell">MS
									<input className="value" disabled value={x.Router.MS} />
								</div>
								<div className="cell">Availability(slots)
									<input className="value" disabled value={x.Router.AvailableSlots} />
								</div>
							</>
						)
					}
				})}

			</div >
		)
	}


	return (
		<div className="connection-wrapper"  >

			<div className="search-wrapper">
				<MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
				<input type="text" className="search" onChange={(e) => setFilter(e.target.value)} placeholder="Search .."></input>
			</div>

			<div className="connections-flex">
				{cons.map((c) => {
					return RenderConnection(c)
				})}
			</div>
		</div >
	);
}

export default Connections;
