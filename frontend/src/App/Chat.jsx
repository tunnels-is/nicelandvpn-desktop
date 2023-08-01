import React, { useEffect, useRef, useState } from "react";
import STORE from "../store";

const Chat = (props) => {

    const [socket, setSocket] = useState(undefined)
    const [history, setHistory] = useState([])
    const [channel, setChannel] = useState("")
    const [agent, setAgent] = useState("")
    const inputRef = useRef(null);

    const AddLineToHistory = (line) => {
        let h = STORE.Cache.GetObject("history")
        let nh = [...h]
        nh.push(line)
        STORE.Cache.SetObject("history", nh)
        setHistory(nh)
    }

    const handleSend = (event) => {
        if (event.key === 'Enter') {
            console.log('SEND LINE:', event.target.value)
            console.dir(event)
            if (channel === "") {
                return
            }
            let MSG = {
                Channel: channel,
                Agent: agent,
                TEXT: event.target.value,
                TO: "agent",
            }

            socket.send(JSON.stringify(MSG))
            AddLineToHistory(MSG)
            inputRef.current.value = "";
        }
    }

    const OpenChatSocket = (email) => {
        let ws = undefined
        try {
            ws = new WebSocket("ws://localhost:1333/chat")
            setSocket(ws)
        } catch (error) {
            console.dir(error)
            return
        }


        if (ws === undefined) {
            console.log("COULD NOT OPEN CHAT")
            return
        }

        ws.onopen = function () {
            console.log('Connected')
            ws.send(email);
        }

        ws.onmessage = function (evt) {
            let msg = JSON.parse(evt.data)
            if (msg.CONTROL === "start-support") {
                setAgent(msg.Agent)
                AddLineToHistory({ TEXT: msg.Agent + " has joined the chat", TO: "channel", Agent: "Notification" })
                return
            }
            AddLineToHistory(msg)
        }

    }
    const RestartChat = () => {
        AddLineToHistory({ TEXT: "Support chat reconnected", TO: "channel", Agent: "Notification" })
        if (channel === "") {
            return
        }
        OpenChatSocket(channel)
    }

    const StartChat = () => {
        STORE.Cache.SetObject("history", [])
        AddLineToHistory({ TEXT: "Welcome to the support chat, please wait a moment for a customer service agent. If the chat disconnects, you can press 'reconnect' to get back in", TO: "channel", Agent: "Notification" })
        if (channel === "") {
            return
        }
        OpenChatSocket(channel)

    }

    useEffect(() => {
        let user = STORE.Cache.GetObject("user")
        if (user) {
            setChannel(user.Email)
        }

        let x = STORE.Cache.GetObject("history")
        if (x && x.length > 1) {
            setHistory(x)
        }

    }, [])

    let rh = [...history]
    rh.reverse()

    let chatEnabled = STORE.Cache.GetBool("live-chat")
    if (!chatEnabled) {
        return (<></>)
    } else {
        return (
            <div className="support-chat">

                <div className="chat-window">
                    {rh.map((line, index) => {
                        if (line.TO === "agent") {
                            return (
                                <div key={index} className="line">
                                    <div className="from-self">
                                        <div className="channel-name">
                                            {line.Channel}
                                        </div>
                                        {line.TEXT}
                                    </div>
                                </div>
                            )
                        } else if (line.TO === "channel") {
                            return (
                                <div key={index} className="line from-agent">
                                    <div className="from-agent">
                                        <div className="agent-name">
                                            {line.Agent}
                                        </div>
                                        {line.TEXT}
                                    </div>
                                </div>
                            )
                        }
                    })}
                </div>

                <input ref={inputRef} className="input" type="text" placeholder="..." onKeyDown={handleSend}></input>

                <div className="support-buttons">
                    <div className="start" onClick={StartChat}>New Chat</div>
                    <div className="reconnect" onClick={RestartChat}>Reconnect</div>
                </div>
            </div >
        )
    }

}

export default Chat;