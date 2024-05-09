import { useEffect, useRef, useState, useCallback } from "react";
import Button from "@mui/material/Button";
import "./style.css";

function useInterval(callback, delay) {
  const savedCallback = useRef();

  // Remember the latest callback.
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  // Set up the interval.
  useEffect(() => {
    function tick() {
      savedCallback.current();
    }
    if (delay !== null) {
      let id = setInterval(tick, delay);
      return () => clearInterval(id);
    }
  }, [delay]);
}

export default function SecondFactor({ next, email }) {
  const websocket = useRef(null);

  let remoteDescriptionSet = false;
  const candidateQueue = [];

  function flushCandidateQueue() {
    while (candidateQueue.length > 0) {
      const candidate = candidateQueue.pop();
      peerConnection.current.addIceCandidate(candidate);
    }
  }

  const configuration = {
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
  };

  async function startConnection() {
    console.log("Starting connection...");
    const offer = await peerConnection.current.createOffer();
    await peerConnection.current.setLocalDescription(offer);
    sendOffer();
  }

  function sendOffer() {
    const message = {
      command: "IceOffer",
      payload: peerConnection.current.localDescription,
    };
    websocket.current.send(JSON.stringify(message));
    console.log("Offer sent.");
  }

  // document.querySelector("#showVideo").addEventListener("click", e => initialize(e))

  async function attachVideoStream() {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: false,
      video: { width: 1280, height: 720 },
    });
    window.stream = stream;
    camVideo.current.srcObject = stream;
    peerConnection.current.addTrack(stream.getVideoTracks()[0], stream);
  }

  function closeInstruction(e) {
    setInstructions(false);

    const message = {
      command: "readyForBandColor",
      payload: { email },
    };
    websocket.current.send(JSON.stringify(message));

    readyTimeout();
  }

  async function readyTimeout() {
    await sleep(10000);
    peerConnection.current.close();
    console.log("connection Closed");
  }

  async function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  const [backgroundCol, setBackgroundCol] = useState();
  const [bandCol, setBandCol] = useState();
  const [bandPos, setBandPos] = useState();
  const [instructions, setInstructions] = useState(true);
  const [displayData, setDisplayData] = useState(null);
  const [display, setDisplay] = useState(false);
  const camVideo = useRef();
  const [nextData, setNextData] = useState<any>([]);
  const [colorIndex, setColorIndex] = useState(0);

  async function confirmColorChange() {
    console.log("Acknowledging color change");
    let date = new Date();
    const information = {
      timestamp: date.toISOString(),
      backgroundColor: backgroundCol,
      stripColor: bandCol,
      stripPosition: bandPos,
      index: colorIndex,
    };

    const message = {
      command: "colorCommandAck",
      payload: { information },
    };

    websocket.current.send(JSON.stringify(message));
  }

  // const updateColor = useCallback(() => {

  //   console.log("in update color", nextData);
  //   if (display && nextData.length !== 0) {
  //     setBackgroundCol(nextData[0].backgroundCol);
  //     setBandPos(nextData[0].stripPos);
  //     setBandCol(nextData[0].stripCol);
  //     confirmColorChange();
  //     setNextData((data: any) => {
  //       const clone = [...data];
  //       clone.shift();
  //       return clone;
  //     });
  //     setColorIndex(colorIndex + 1);
  //   }
  // }, [display, nextData, colorIndex, confirmColorChange]);

  useInterval(() => {
    if (display && nextData.length !== 0) {
      setBackgroundCol(nextData[0].backgroundColor);
      setBandPos(nextData[0].stripPosition+"%");
      setBandCol(nextData[0].stripColor);
      confirmColorChange();
      setNextData((data: any) => {
        const clone = [...data];
        clone.shift();
        return clone;
      });
      setColorIndex(colorIndex + 1);
    }
  }, 500);

  const peerConnection = useRef(null);
  useEffect(() => {
    if (!websocket.current) {
      websocket.current = new WebSocket("ws://localhost:8080/ws");

      websocket.current.onclose = e => {
        console.log("WebSocket connection closed.");
      };

      websocket.current.onopen = e => {
        console.log("WebSocket connection established.");
        startConnection();
      };

      websocket.current.onmessage = async e => {
        const message = JSON.parse(e.data);
        console.log("Received message:", message);

        if (message.command === "IceAnswer") {
          const remoteDesc = new RTCSessionDescription(message.payload);
          console.log("Setting remote description...");
          await peerConnection.current.setRemoteDescription(remoteDesc);
          remoteDescriptionSet = true;
          flushCandidateQueue();
        } else if (message.command === "IceCandidate") {
          const candidate = new RTCIceCandidate(message.payload);
          if (remoteDescriptionSet) {
            await peerConnection.current.addIceCandidate(candidate);
          } else {
            candidateQueue.push(candidate);
          }
        } else if (message.command === "setBandColor") {
          if (display == false) {
            await attachVideoStream();
            setDisplay(true);
          }

          // console.log("Setting band color...");
          setNextData(data => {
            let clone = structuredClone(data);
            // console.log("pushing", message.payload);
            clone.push(message.payload);

            // console.log(clone);
            return clone;
          });
        } else if(message.command === "authenticationResult"){
          if(message.payload.success){
            next();
          }
        }
      };
    }

    // Initialize the peer connection when the component mounts
    if (!peerConnection.current) {
      peerConnection.current = new RTCPeerConnection(configuration);

      let dataChannel =
        peerConnection.current.createDataChannel("myDataChannel");

      peerConnection.current.onicecandidate = event => {
        if (event.candidate) {
          console.log("Sending new ICE candidate...");
          websocket.current.send(
            JSON.stringify({
              command: "IceCandidate",
              payload: event.candidate.toJSON(),
            }),
          );
        } else {
          console.log("ICE gathering complete.");
        }
      };

      peerConnection.current.onconnectionstatechange = event => {
        console.log(
          "Connection state change:",
          peerConnection.current.connectionState,
        );
      };

      // Additional setup like handling incoming data channels or streams
    }
  }, []);

  return (
    <>
      <video
        id="gum-local"
        playsInline
        autoPlay
        ref={camVideo}
        style={{
          position: "absolute",
          left: 0,
          right: 0,
          top: 0,
          bottom: 0,
          margin: "auto",
          width: "100vw",
        }}
      />
      <div
        style={{
          backgroundColor: backgroundCol,
          width: "100vw",
          height: "100vw",
          opacity: "70%",
        }}>
        <div
          style={{
            backgroundColor: bandCol,
            width: "100vw",
            height: "20%",
            position: "absolute",
            top: bandPos,
          }}></div>
      </div>
      <img
        src="oval.png"
        style={{
          position: "absolute",
          top: 0,
          bottom: 0,
          left: 0,
          right: 0,
          margin: "auto",
          height: "70%",
        }}
      />

      {instructions && (
        <div className="instructionPopup">
          <div className="instructions">
            <h1 className="instructionHead">
              Instructions for Facial Authentication
            </h1>
            <ul>
              <li>Ensure your face is brightly lit</li>
              <li>Position your face according to the outline</li>
              <li>Hold Still until flashing is complete</li>
            </ul>
            <Button className="instructionButton" onClick={closeInstruction}>
              I'm Ready
            </Button>
          </div>
        </div>
      )}
    </>
  );
}
