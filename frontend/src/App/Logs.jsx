import React, { useEffect, useState } from "react";

import { GetLogs } from "../../wailsjs/go/main/Service";
import { MagnifyingGlassIcon } from "@radix-ui/react-icons";

const Logs = (props) => {

  const [filter, setFilter] = useState("");
  const [data, setData] = useState({})
  const [timer, setTimer] = useState(1)
  const [logCount, setlogCount] = useState(0)
  const [timeoutMS, setTimeoutMS] = useState(500)

  useEffect(() => {

    const to = setTimeout(async () => {

      await GetLogs(logCount, filter).then((x) => {
        if (x.Data) {
          setData(x.Data)
          setlogCount(x.Data.Content.length)
        }

      }).catch((e) => {
        console.dir(e)
      })

      let t = timer + 1
      if (t > 10000) {
        t = 1
      }

      setTimer(t)

    }, timeoutMS)

    return () => { clearTimeout(to); }

  }, [timer])

  let logLines = []
  if (filter !== "") {

    data.Content.forEach((line) => {
      let lowerLine = line.toLowerCase()
      if (lowerLine.includes(filter)) {
        logLines.push(line)
      }
    });

  } else {
    if (data.Content) {
      logLines = data.Content
    }
  }

  return (
    <div className="logs-wrapper">


      <div className="search-wrapper">
        <MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
        <input type="text" className="search" onChange={(e) => setFilter(e.target.value)} placeholder="Search .."></input>
      </div>

      <div className="logs-window custom-scrollbar">

        {logLines.map((content, index) => {
          return (
            <div className="line">
              <label className="time">{data.Time[index]}</label>
              {" || "}
              <label className="orange">{data.Function[index]}</label>
              {" || "}
              <label className={"" + data.Color[index]}>{content}</label>
            </div >
          )
        })}

      </div>
    </div >
  );


}

export default Logs;
