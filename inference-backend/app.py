from flask import Flask, request, jsonify
from deepface import DeepFace
from PIL import Image
import numpy as np
import json
import base64
import os
from inference import process, calculate, infer
import pickle
import csv

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
        embedding_objs = DeepFace.represent(img_path=img_array, model_name='VGG-Face', enforce_detection=False)
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
    json_data = request.json

    if not json_data or 'sessionId' not in json_data:
        return jsonify({'error': 'No sessionId provided'})

    session_id = json_data['sessionId']

    video_path = os.path.join('video', session_id + '.mp4')
    csv_path = os.path.join('csv', session_id + '.csv')
    frames_path = os.path.join('frames', session_id)
    lr_model_path = "/path/to/your/lr_model.pkl"

    if not os.path.exists(lr_model_path):
        return jsonify({'error': 'LR model does not exist'})

    try:
        if os.path.exists(video_path) and os.path.exists(csv_path):
            # Load the model from the file
            with open(lr_model_path, 'rb') as f:
                lr_model = pickle.load(f)

            # split video into frames
            process.split_video(video_path, frames_path)
            # crop frames
            process.crop_frames(frames_path, frames_path)
            # find color changes
            color_changes = calculate.color_change(csv_path)
            # liveness detection
            success = infer.predict_liveliness(csv_path, frames_path, color_changes, lr_model)

            if not success:
                return jsonify({'authenticated': False, 'embedding': None})
        else:
            return jsonify({'error': 'One or both files not found'})

    except Exception as e:
        return jsonify({'error': f'An error occurred: {str(e)}'})

    try:
        random_image = process.get_random_frame(video_path)
        embedding = infer.generate_embedding(random_image)
    except Exception as e:
        return jsonify({'error': f'An error occurred: {str(e)}'})

    embedding_json = json.dumps(embedding)  # Serialize the embedding list to JSON
    encoded_embedding = base64.b64encode(embedding_json.encode('utf-8'))  # Encode the JSON string using Base64

    return jsonify({'authenticated': True,
                    'embedding': encoded_embedding.decode('utf-8')})

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
