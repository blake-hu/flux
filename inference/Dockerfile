FROM python:3.9-slim

RUN apt-get update && apt-get install -y pkg-config ffmpeg
RUN apt-get update && \
    apt-get install -y build-essential libhdf5-dev  \
    libgl1-mesa-glx \
    libglib2.0-0 \
    libsm6 \
    libxrender1 \
    libxext6

WORKDIR /deploy/inference

COPY ./inference/requirements.txt requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY ./inference .

CMD ["python", "app.py"]
