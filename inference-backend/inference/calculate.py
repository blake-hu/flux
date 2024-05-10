import pandas as pd


def eqn_2(pixel, color1, color2, illuminance):
    '''
    #INPUTS:
        pixel: a single pixel with all 3 color channels
        color1: background color being shown on screen
        color2: primary color being shown on screen (band)
        E: illuminance for all 3 channels

    Confirm that I{c1}/I{c2} = E{c1}/E{c2} (where c1 and c2 are the 2 colors being shown on the screen)
    '''

    i_fraction = pixel[color1]/(pixel[color2]+1)
    e_fraction = illuminance[color1]/illuminance[color2]
    epsilon = 0.01
    return i_fraction - e_fraction <= epsilon


def verify_eqn2(color1, color2, img):
    # Apply Eqn 2 on every pixel between response of lighting challenge and background challenge
    count = 0
    illuminance = [0, 0, 0]
    illuminance[int(color1)] = 256
    illuminance[int(color2)] = 256
    for r in range(img.shape[0]):
        for c in range(img.shape[1]):
            consistent = eqn_2(img[r][c][:], color1, color2, illuminance)
            if not consistent:
                count += 1
                # print("Not consistent!")
                # return

    return count / (img.shape[0]*img.shape[1])


def roi(t_u, ct_k, ct_frame, image):
    '''
    INPUTS:
        t_u = time that this color started
        u = top of band
        ct_k = start time to exposure the first column of k-th capture frame
            --> find w/ firstImg
        ct_frame = exposure time of one captured frame
            --> average time of each frame ? maybe can calculate w/ dict.
        image = first image whose recording period covers t_u
    '''
    cols = image.shape[1]
    a = cols * (t_u - ct_k)/ct_frame
    b = a + 0.2*image.shape[1]
    return [a, b]


def map_colors(color):
    if color == "Red":
        return 0
    if color == "Green":
        return 1
    else:
        return 2


def remove_percent(s):
    # return float("0." + s[:-1])
    return float(s/100)


def color_change(color_change_csv):
    color_changes = pd.read_csv(color_change_csv)
    color_changes.iloc[:, 0] = color_changes.iloc[:, 0].apply(map_colors)
    color_changes.iloc[:, 1] = color_changes.iloc[:, 1].apply(map_colors)
    color_changes.iloc[:, 2] = color_changes.iloc[:, 2].apply(remove_percent)
    color_changes.iloc[:, 3] = color_changes.iloc[:, 3] - color_changes.iloc[0, 3]
    return color_changes
