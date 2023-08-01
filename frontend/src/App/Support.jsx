import React from "react";

import { OpenURL, CopyToClipboard } from "../../wailsjs/go/main/App";
import { BoxIcon } from "@radix-ui/react-icons";

const Feed = () => {
    return (
        <div className="support-wrapper">

            {/* <div className="info">
                <div className="icon"></div>
                <div className="title">Contact Support</div>
                <div className="text">You can contact us via email, discord, slack, reddit or twitter</div>

                <div className="emails">
                    <div className="email">
                        support@nicelandvpn.is <br />
                    </div>
                </div>
            </div> */}

            <div className="social-wrapper">
                <div className="image-panel " >
                    <a href="mailto: support@nicelandvpn.is" className="text email">
                        EMAIL
                    </a>
                    <div className="copy" onClick={() => CopyToClipboard("support@nicelandvpn.is")}>
                        Copy Link
                    </div>
                </div>

                <div className="image-panel telegram" >
                    <div className="text" onClick={() => OpenURL("https://t.me/+hTRZ3W3YyuQwZGFk")}>TELEGRAM</div>
                    <div className="copy" onClick={() => CopyToClipboard("https://t.me/+hTRZ3W3YyuQwZGFk")}>
                        Copy Link
                    </div>
                </div>

                <div className="image-panel element" >
                    <div className="text" onClick={() => OpenURL("https://matrix.to/#/#nicelandvpn:matrix.org")} >ELEMENT</div>
                    <div className="copy" onClick={() => CopyToClipboard("https://matrix.to/#/#nicelandvpn:matrix.org")}>
                        Copy Link
                    </div>
                </div>

                <div className="image-panel slack" >
                    <div className="text" onClick={() => OpenURL("https://join.slack.com/t/nicelandvpn/shared_invite/zt-1rfv4ks6d-A5lLr9W4FdjEzlmZXwrMzw")}>SLACK</div>
                    <div className="copy" onClick={() => CopyToClipboard("https://join.slack.com/t/nicelandvpn/shared_invite/zt-1rfv4ks6d-A5lLr9W4FdjEzlmZXwrMzw")}>
                        Copy Link
                    </div>
                </div>

                <div className="image-panel discord" >
                    <div className="text" onClick={() => OpenURL("https://discord.gg/7Ts3PCnCd9")}>DISCORD</div>
                    <div className="copy" onClick={() => CopyToClipboard("https://discord.gg/7Ts3PCnCd9")}>
                        Copy Link
                    </div>
                </div>

                <div className="image-panel reddit" >
                    <div className="text" onClick={() => OpenURL("https://www.reddit.com/r/nicelandvpn")}>REDDIT</div>
                    <div className="copy" onClick={() => CopyToClipboard("https://www.reddit.com/r/nicelandvpn")}>
                        Copy Link
                    </div>
                </div>

                <div className="image-panel twitter" >
                    <div className="text" onClick={() => OpenURL("https://www.twitter.com/nicelandvpn")}>TWITTER</div>
                    <div className="copy" onClick={() => CopyToClipboard("https://www.twitter.com/nicelandvpn")}>
                        Copy Link
                    </div>
                </div>


            </div>

        </div >
    )
}

export default Feed;