from flask import Flask, request, jsonify
from deepface import DeepFace
from PIL import Image
import numpy as np

app = Flask(__name__)

# Define a route to accept POST requests with image data
@app.route('/generate_embedding', methods=['POST'])
def generate_embedding():
    # Check if the request contains an image file
    if 'image' not in request.files:
        return jsonify({'error': 'No file part'})

    # Read the image file from the request
    image_file = request.files['image']

    # Process the image and generate embeddings
    try:
        img = Image.open(image_file)
        img_array = np.array(img).astype(int)

        # Detect face and generate embeddings
        embedding_objs = DeepFace.represent(img_path=img_array, model_name='Facenet', enforce_detection=False)

        # Extract the embeddings
        embedding = embedding_objs[0]["embedding"]
        print(embedding)

    except Exception as e:
        return jsonify({'error': str(e)})

    return jsonify({'embedding': embedding})

if __name__ == '__main__':
    app.run(debug=True)
