from flask import Flask, request, jsonify
from deepface import DeepFace

app = Flask(__name__)

# Define a route to accept POST requests with image data
@app.route('/generate_embedding', methods=['POST'])
def generate_embedding():
    # Check if the request contains an image file
    if 'file' not in request.files:
        return jsonify({'error': 'No file part'})

    # Read the image file from the request
    image_file = request.files['file']

    # Process the image and generate embeddings
    try:
        # Detect face and generate embeddings
        result = DeepFace.represent(image_file, model_name='Facenet', enforce_detection=True)

        # Extract the embeddings
        embedding = result['embedding']
    except Exception as e:
        return jsonify({'error': str(e)})

    return jsonify({'embedding': embedding.tolist()})

if __name__ == '__main__':
    app.run(debug=True)
