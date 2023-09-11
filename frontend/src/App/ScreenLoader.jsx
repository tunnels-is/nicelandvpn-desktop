import React, { useEffect, useState } from "react";

import { GetLoadingLogs } from "../../wailsjs/go/main/Service";

const ScreenLoader = (props) => {
  const [lines, setLines] = useState([])
  const [timer, setTimer] = useState(0)

  useEffect(() => {
    let logs = undefined
    let tag = ""

    const to = setTimeout(async () => {
      if (props.loading?.logTag) {
        tag = props.loading.logTag.toLowerCase()
        await GetLoadingLogs(tag).then((x) => {
          logs = x.Data.Lines
        }).catch((e) => {
          console.dir(e)
        })
      }

      if (logs && logs.length > 0) {
        let xlines = []
        logs.map(l => {
          if (l !== "") {
            let x1 = ""
            try {
              let x = l.split("||")
              x1 = x[0].split(" ")[1]
              if (x.length > 3) {
                xlines.push(<div className="line">
                  <div className="timestamp">{x1}</div >
                  <div className="seperator">||</div >
                  <div className="error">{x[2]}</div >
                  <div className="seperator">||</div >
                  <div className="data">{x[3]}</div >
                </div>)
              } else {
                xlines.push(<div className="line">
                  <div className="timestamp">{x1}</div >
                  <div className="seperator">||</div >
                  <div className="data">{x[2]}</div >
                </div>)
              }
            } catch (error) {
              xlines.push(<div className="line">{l}</div>)
            }
          }
        })

        if (xlines.length > 0) {
          // console.log("setting logs lines")
          // console.dir(xlines)
          setLines(xlines)
        }
      }

      let t = timer + 1
      if (t > 10000) {
        t = 1
      }
      setTimer(t)

    }, 200)

    return () => { clearTimeout(to); }

  }, [timer])

  return (
    <div className="screen-loader" >
      {props.loading?.msg &&
        <div className="title">
          {props.loading.msg}
        </div>
      }

      {props.loading?.subtitle &&
        <div className="subtitle">{props.loading.subtitle}</div>
      }

      {(props.loading?.includeLogs && lines.length > 0) &&
        <div className="scroll-frame">
          {lines}
        </div>
      }
    </div>

  )
}

export default ScreenLoader;