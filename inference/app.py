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
import joblib
import ffmpeg

app = Flask(__name__)
vgg_model = DeepFace.build_model('VGG-Face')

@app.route('/liveness_detection', methods=['POST'])
def liveness_detection():
    json_data = request.json

    if not json_data or 'sessionId' not in json_data:
        return jsonify({'error': 'No sessionId provided'})

    if 'start_offset' not in json_data:
        return jsonify({'error': 'No start_offset provided'})

    session_id = json_data['sessionId']

    video_path = os.path.join('./files/video', session_id + '.ivf')
    cropped_video_path = os.path.join('./files/video', session_id + '.mp4')
    csv_path = os.path.join('./files/csv', session_id + '.csv')
    frames_path = os.path.join('frames/video', session_id)
    cropped_path = os.path.join('frames/cropped', session_id)
    lr_model_path = "model.joblib"  # TODO (change model path)

    # integer in micro seconds 10^-6
    start_offset = int(json_data['start_offset'])

    if not os.path.exists(lr_model_path):
        return jsonify({'error': 'LR model does not exist'})

    try:
        if os.path.exists(video_path) and os.path.exists(csv_path):

            # convert ivf to mp4 and crop delete_duration off the beginning of the video
            ffmpeg.input(video_path).output(cropped_video_path).run()

            # Load the model from the file
            lr_model = joblib.load(lr_model_path)

            # split video into frames
            print('splitting video')
            process.split_video(cropped_video_path,
                                frames_path, start_offset)
            print('cropping frames')
            # crop frames
            success = process.crop_frames(cropped_path, frames_path)
            if not success:
                # change return
                return jsonify({'error': 'Could not detect face'})

            # find color changes
            print('calculating color changes')
            color_changes = calculate.color_change(csv_path)
            # liveness detection
            print('predicting liveness')
            success = infer.predict_liveliness(
                csv_path, cropped_path, color_changes, lr_model)

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

    # Serialize the embedding list to JSON
    embedding_json = json.dumps(embedding)
    encoded_embedding = base64.b64encode(embedding_json.encode(
        'utf-8'))  # Encode the JSON string using Base64

    return jsonify({'authenticated': True,
                    'embedding': encoded_embedding.decode('utf-8')})


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
