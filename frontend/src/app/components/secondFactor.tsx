import { useEffect, useRef, useState, useCallback } from 'react';
import Button from '@mui/material/Button';
import "./style.css"
import next from 'next';


export default function SecondFactor({ back ,email}) {
  const websocket = new WebSocket("ws://localhost:8080/ws");
  websocket.onopen = e => {
    console.log("WebSocket connection established.");
    startConnection();
  };

  let remoteDescriptionSet = false;
  const candidateQueue = [];

  websocket.onmessage = async e => {
    const message = JSON.parse(e.data);
    console.log("Received message:", message);

    if (message.command === "IceAnswer") {
      const remoteDesc = new RTCSessionDescription(message.payload);
      console.log("Setting remote description...");
      await peerConnection.setRemoteDescription(remoteDesc);
      remoteDescriptionSet = true;
      flushCandidateQueue();
    } else if (message.command === "IceCandidate") {
      const candidate = new RTCIceCandidate(message.payload);
      if (remoteDescriptionSet) {
        await peerConnection.addIceCandidate(candidate);
      } else {
        candidateQueue.push(candidate);
      }
    } else if (message.command === "setBandColor") {
      if (display == false) {
        attachVideoStream(camVideoStream)
      }
      setDisplay(true)
      setNextData(data => {
        let clone = structuredClone(data)
        clone.push(message.payload)
        return clone
      })
    }

  };


  function flushCandidateQueue() {
    while (candidateQueue.length > 0) {
      const candidate = candidateQueue.pop();
      peerConnection.addIceCandidate(candidate);
    }
  }

  websocket.onerror = error => {
    console.error("WebSocket Error:", error);
  };

  websocket.onclose = event => {
    console.log("WebSocket Closed:", event.reason);
  };

  const configuration = {
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
  };
  // const dataChannel = peerConnection.createDataChannel("myDataChannel");



  async function startConnection() {
    console.log("Starting connection...");
    const offer = await peerConnection.createOffer();
    await peerConnection.setLocalDescription(offer);
    sendOffer();
  }

  function sendOffer() {
    const message = {
      command: "IceOffer",
      payload: peerConnection.localDescription,
    };
    websocket.send(JSON.stringify(message));
    console.log("Offer sent.");
  }

  

  // document.querySelector("#showVideo").addEventListener("click", e => initialize(e))

  async function initialize() {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: false,
      video: { width: 1280, height: 720 }

    });
    setCamVideoStream(stream)
    peerConnection.onicecandidate = event => {
      if (event.candidate) {
        console.log("Sending new ICE candidate...");
        websocket.send(
          JSON.stringify({
            command: "IceCandidate",
            payload: event.candidate.toJSON(),
          }),
        );
      } else {
        console.log("ICE gathering complete.");
      }
    };

    peerConnection.onconnectionstatechange = event => {
      console.log("Connection state change:", peerConnection.connectionState);
    };
  }

  function attachVideoStream(stream) {
    const videoElement = document.querySelector("video");
    window.stream = stream;
    camVideo.current.srcObject = stream;
    peerConnection.addTrack(stream.getVideoTracks()[0], stream);
  }

  function closeInstruction(e) {
    setInstructions(false);

    const message = {
      command: "readyForBandColor",
      payload: {email}
    };
    websocket.send(JSON.stringify(message));

    readyTimeout();
  }

  async function readyTimeout() {
    await sleep(10000);
    peerConnection.close();
    console.log("connection Closed");
  }

  async function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));

  }

  const [backgroundCol, setBackgroundCol] = useState()
  const [bandCol, setBandCol] = useState()
  const [bandPos, setBandPos] = useState()
  const [instructions, setInstructions] = useState(true)
  const [displayData, setDisplayData] = useState(null)
  const [display, setDisplay] = useState(false)
  const camVideo = useRef()
  const [nextData, setNextData] = useState([])
  const [colorIndex, setColorIndex] = useState(0)
  const [camVideoStream, setCamVideoStream] = useState()
  changeColor()




  async function confirmColorChange() {
    console.log("Acknowledging color change");
    let date = new Date()
    const information = {
      timestamp: date.toISOString(),
      backgroundColor: backgroundCol,
      stripColor: bandCol,
      stripPosition: bandPos,
    }

    const message = {
      command: "ColorCommandAck",
      payload: {information},
    };

    websocket.send(JSON.stringify(message));
  }

  
  async function changeColor() {
    let intervalId;
  
    function updateColor() {
      if (display && nextData.length !== 0) {
        setBackgroundCol(nextData[0].backgroundCol);
        setBandPos(nextData[0].stripPos);
        setBandCol(nextData[0].stripCol);
        confirmColorChange();
        setNextData(data => {
          const clone = [...data];
          clone.shift();
          return clone;
        });
        setColorIndex(colorIndex + 1);
      }
    }
  
    intervalId = setInterval(updateColor, 500);
  
    return () => clearInterval(intervalId); // Cleanup function
  }

  const peerConnection = useRef(null);
  useEffect(() => {

    

    // Initialize the peer connection when the component mounts
    if (!peerConnection.current) {
        peerConnection.current = new RTCPeerConnection();

        initialize()
        
        
        // Additional setup like handling incoming data channels or streams
    }
        
        
  },[])
  

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
