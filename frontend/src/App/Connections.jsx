import { Navigate } from "react-router-dom";
import React, { useState } from "react";


import STORE from "../store";
import { DesktopIcon, MagnifyingGlassIcon, EnterIcon } from "@radix-ui/react-icons";


var connections = [
	{
		IPv4Address: "10.4.3.2",
		IPv6Address: "",
		Name: "Niceland",
		MTU: 1500,
		TxQueueLen: 3000,
		Persistent: true,
		Routes: ["default"],
		Router: "*",
		Node: "*",
		AutoReconnect: false,
		DNS: ["1.1.1.1"],
	},
	{
		IPv4Address: "10.4.3.3",
		IPv6Address: "",
		Name: "cloud-edge-router",
		MTU: 1500,
		TxQueueLen: 3000,
		Persistent: true,
		Routes: ["172.17.1.0/24", "172.18.1.0/24"],
		Router: "london-01",
		Node: "edge-router",
		AutoReconnect: true,
		DNS: ["172.19.1.1"],
	}
]

const Connections = (props) => {

	const [filter, setFilter] = useState("");

	STORE.Cache.SetObject("connections", connections)
	props.state.Connections = connections

	if (props?.state?.Connections) {

		if (filter && filter !== "") {
			props.state.Connections.map(r => {

				let filterMatch = false
				if (r.Tag.includes(filter)) {
					filterMatch = true
				}

				if (filterMatch) {
					connections.push(r)
				}
			})

		} else {
			connections = props.state.Connections
		}

	}

	const RenderConnection = (c) => {
		return (
			<div className="connection" key={"x"} onClick={() => console.dir("clicked")}>
				<h3>{c.Name}</h3>
				<h3>{c.IPv4Address}</h3>
				<h3>{c.IPv6Address}</h3>
				<h3>{c.MTU}</h3>
				<h3>{c.TxQueueLen}</h3>
				<h3>{c.Persistent}</h3>
				<h3>{c.Router}</h3>
				<h3>{c.Node}</h3>
				<h3>{c.Persistent}</h3>
				<h3>{c.AutoReconnect}</h3>
				{c.Routes.map((r) => {
					return (<h4>{r}</h4>)
				})}
				{c.DNS.map((r) => {
					return (<h4>{r}</h4>)
				})}
			</div>
		)
	}


	return (
		<div className="connection-wrapper"  >

			<div className="search-wrapper">
				<MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
				<input type="text" className="search" onChange={(e) => setFilter(e.target.value)} placeholder="Search .."></input>
			</div>

			{connections.map((c) => {
				console.log("one!")
				return RenderConnection(c)
			})}

		</div >
	);
}

export default Connections;
