import React, { useEffect, useState } from "react";

const Browser = (props) => {

    const [url, setUrl] = useState("https://nicelandvpn.com")

    const changeURL = (event) => {
        if (event.key === 'Enter') {
            setUrl(event.target.value)
        }
    }
    return (
        <>
            <input style={{ marginTop: "30px" }} type="text" onKeyDown={changeURL} />
            <embed type="text/html" src={url} style={{ height: "100%", width: "100%" }} />
            {/* <iframe style={{ height: "100%", width: "100%" }} src={url}></iframe> */}
        </>
    )
}

export default Browser;