from flask import Flask, request, jsonify
from deepface import DeepFace
from PIL import Image
import numpy as np
import json
import base64

app = Flask(__name__)

@app.route('/generate_embedding', methods=['POST'])
def generate_embedding():

    if 'image' not in request.files:
        return jsonify({'error': 'No file part'})

    image_file = request.files['image']

    try:
        img = Image.open(image_file)
        img_array = np.array(img).astype(int)
        embedding_objs = DeepFace.represent(img_path=img_array, model_name='VGG-Face', enforce_detection=False)
        embedding = embedding_objs[0]["embedding"]

    except Exception as e:
        return jsonify({'error': str(e)})

    # Serialize the embedding list to JSON
    embedding_json = json.dumps(embedding)

    # Encsode the JSON string using Base64
    encoded_embedding = base64.b64encode(embedding_json.encode('utf-8'))
    
    return jsonify({'encoded_embedding': encoded_embedding.decode('utf-8')})


if __name__ == '__main__':
    app.run(debug=True)
