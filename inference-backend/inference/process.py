from . import *

def split_video(input_video, output_directory1):
    # takes input video and output directory
    # Split video into frames in the videoframes directory
    # Runtime: ~60 seconds

    # Delete directory if it already exists and make a new one
    if os.path.exists(output_directory1):
        shutil.rmtree(output_directory1)
    os.makedirs(output_directory1)

    vidcap = cv2.VideoCapture(input_video)
    fps = vidcap.get(cv2.CAP_PROP_FPS)  # get fps
    frametime_list = []
    print("FPS", fps)

    if vidcap.isOpened():
        while True:
            ret, frame = vidcap.read()

            if not ret:
                break

            # Get the timestamp of the current frame in milliseconds
            timestamp = round(vidcap.get(cv2.CAP_PROP_POS_MSEC))
            frametime_list.append(timestamp)

            # Save each frame as an image in the output folder
            frame_name = f"frame_{timestamp}.jpg"
            frame_path = os.path.join(output_directory1, frame_name)
            cv2.imwrite(frame_path, frame)

        vidcap.release()

        # frametime_list = np.array(frametime_list)

        print(f"Video frames saved in '{output_directory1}'")

def crop_frames(frames, output_directory1):
    # CONVERT FRAMES INTO CROPPED FRAMES
    if os.path.exists(frames):
        shutil.rmtree(frames)
    os.makedirs(frames)

    for filename in sorted(os.listdir(output_directory1)):
        frame_path = os.path.join(output_directory1, filename)

        if os.path.isfile(frame_path):  # Check if it's a file (not a subdirectory)
            # face detection and alignment
            img = cv2.imread(frame_path)

            try:  # skip frame if it can't detect face
                face_objs = DeepFace.extract_faces(img_path=frame_path,
                                                   target_size=(224, 224)
                                                   )
            except:
                print("Couldn't detect face in frame " + frame_path)

            face = face_objs[0]['facial_area']
            x, y, w, h = face['x'], face['y'], face['w'], face['h']

            img = img[y:y + h, x:x + w]

            # Save each frame as an image in the output folder
            # frame_name = f"frame_{frame_count:04d}.jpg"
            frame_path = os.path.join(frames, filename)
            cv2.imwrite(frame_path, img)

def get_random_frame(video_file):
    cap = cv2.VideoCapture(video_file)
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    random_frame_index = random.randint(0, total_frames - 1)

    cap.set(cv2.CAP_PROP_POS_FRAMES, random_frame_index)
    ret, frame = cap.read()
    if ret:
        return frame
    else:
        print("Error: Unable to read frame")
        return None

