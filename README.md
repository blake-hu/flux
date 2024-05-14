# Flux

Flux is a distributed, spoof-resistant facial authentication service.

* **Distributed**: Built on peer-to-peer protocols like WebRTC and WebSocket with a scalable microservice architecture.
* **Spoof-resistant**: Detects face spoofs and deepfakes using camera physics and a regression model, as described in [this paper](https://arxiv.org/abs/1801.01949).
* **Accurate Authentication**: Uses the VGGNet convolutional neural network and a vector database to identify individuals.
* **Hardware-agnostic**: Works on any device with a camera and screen—no specialized hardware needed.

## Components

1. **Frontend**: Next.js website which implements color flashing and video streaming.
2. **Server**: Go server which sends control commands to frontend, processes video streams and manages the vector database.
3. **Inference**: Python backend deploying computer vision models for liveness detection and convolutional neural networks for facial recognition.
4. **Training**: Jupyter notebooks for training liveness detection regression model.

## Installation

To start backend services, place a `model.joblib` file at `/inference/model.joblib`, then run:

```bash
docker compose up --build
```

To start frontend services, ensure that you are on Node 18, then run:

```bash
cd frontend
npm i
npm run dev
```

## Model Training

Note that this repository does not come with a pre-existing dataset or model due to privacy concerns. To train the regression model using your data, please use the Jupyter notebooks in `/training`. Training data can be generated using the frontend module by recording videos of different faces in front of the flashing color bands.
