# import dependencies

import numpy as np
import cv2
from sklearn.linear_model import LinearRegression
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_squared_error as mse
from sklearn import svm, datasets
import pandas as pd
from deepface import DeepFace
import os
import shutil
import random
from PIL import Image
import pickle

_all__ = ['process', 'infer'] # rename inference (same name as module)