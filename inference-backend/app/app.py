from flask import Flask, request, jsonify
from deepface import DeepFace
from PIL import Image
import numpy as np
import json
import base64
import os
from ..inference import process, calculate, inference

app = Flask(__name__)
vgg_model = DeepFace.build_model('VGG-Face')


@app.route('/generate_embedding', methods=['POST'])
def generate_embedding():

    if 'image' not in request.files:
        return jsonify({'error': 'No file part'})

    image_file = request.files['image']

    try:
        img = Image.open(str(image_file))
        img_array = np.array(img).astype(int)
        embedding_objs = DeepFace.represent(
            img_path=img_array, model_name='VGG-Face', enforce_detection=False)
        embedding = embedding_objs[0]["embedding"]

    except Exception as e:
        return jsonify({'error': str(e)})

    # Serialize the embedding list to JSON
    embedding_json = json.dumps(embedding)

    # Encode the JSON string using Base64
    encoded_embedding = base64.b64encode(embedding_json.encode('utf-8'))

    return jsonify({'encoded_embedding': encoded_embedding.decode('utf-8')})


@app.route('/liveness-detection', methods=['POST'])
def liveness_detection():
    if 'requestId' not in request.files:
        return jsonify({'error': 'No requestId'})

    request_id = request.files['requestId']
    video_path = os.path.join('/video/', request_id, '.mp4')
    csv_path = os.path.join('/csv/', request_id, '.csv')
    frames_path = os.path.join('/frames/', request_id)

    try:
        if os.path.exists(video_path) and os.path.exists(csv_path):
            frames_times_list = process.split_video(video_path, frames_path)
            process.crop_frames(frames_times_list, frames_path)
        else:
            # One or both files do not exist
            return jsonify({'error': 'One or both files not found'})
        
    except Exception as e:
        return jsonify({'error': str(e)})


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
