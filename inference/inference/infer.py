from . import *
from .calculate import verify_eqn2, roi


def get_average_vector(roi):
    # Calculate the mean along the height and width (axis 0 and 1), resulting in mean color
    return np.mean(roi, axis=(0, 1))


def lr_fit(rois, targets):
    # predictions = []
    for i in range(len(rois)):
        avg_vector = get_average_vector(rois[i])
        lr = LinearRegression().fit([avg_vector], [targets[i]])

    return lr
    # return predictions


def lr_predict(lr_model, rois):
    predictions = []
    for i in range(len(rois)):
        avg_vector = get_average_vector(rois[i])
        prediction = lr_model.predict([avg_vector])
        predictions.append(prediction)

    return predictions


def find_closest_filename(folder_path, target):
    closest_filename = None
    closest_distance = float('inf')  # Initialize with a large number
    best_ctk = 0

    # Iterate over each file in the directory
    for filename in os.listdir(folder_path):
        if filename.startswith('frame_') and filename.endswith('.jpg'):
            # Extract the number from the filename
            number_part = filename.replace('frame_', '').replace('.jpg', '')
            try:
                number = int(number_part)
                if number <= target:  # tu - ctk >= 30
                    # Calculate the absolute difference from the target
                    distance = abs(number - target)

                    # Update the closest filename if this file is closer
                    if distance < closest_distance:
                        closest_distance = distance
                        closest_filename = filename
                        best_ctk = number
            except ValueError:
                # Handle the case where conversion to int fails
                continue

    return closest_filename, best_ctk


def predict_liveliness(datafile_path, frames, color_changes, lr_model):
    dVals = []
    datafile = pd.read_csv(datafile_path)

    for i in range(len(datafile)):  # for each color change
        color1, color2 = color_changes.iloc[i, 0], color_changes.iloc[i, 1]
        u, startTime = color_changes.iloc[i, 2], color_changes.iloc[i, 3]
        offset = random.randint(2, 6)  # randomly select a ms offset
        t_u = startTime + offset

        filename, ct_k = find_closest_filename(frames, t_u)
        if filename is None:
            continue
        img = cv2.imread(f'{frames}/{filename}')

        exposureTime = 50  # ideally should get file timestamp, if we get frametime list here for the current video file, can get it from there

        # calculate eqn2
        inconsistency = verify_eqn2(color1, color2, img)
        if inconsistency > 0.1:  # STOP HERE IF DOESNT PASS
            continue

        a, b = roi(t_u, startTime, exposureTime, img)
        a = round(a)
        b = round(b)

        imgRows = img.shape[0]
        roi_image = img[:, int(a):int(b)]

        rois = [roi_image]
        y_hat_i = lr_predict(lr_model, rois)
        d_i = y_hat_i[0] - (u + u + imgRows * 0.2) / \
            2  # band is shown on 20% of screen
        dVals.append(d_i)

    mean = np.mean(dVals)
    var = np.var(dVals)
    threshold = -5

    return mean * np.sqrt(var) < np.exp(threshold)


def generate_embedding(image_array):
    # Convert the numpy array to PIL image
    img = Image.fromarray(image_array.astype('uint8'))
    embedding_objs = DeepFace.represent(
        img_path=img, model_name='VGG-Face', enforce_detection=False)
    embedding = embedding_objs[0]["embedding"]

    return embedding
