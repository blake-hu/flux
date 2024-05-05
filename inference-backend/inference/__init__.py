# import dependencies

import numpy as np
import cv2
from sklearn.linear_model import LinearRegression
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_squared_error as mse
import pandas as pd
from deepface import DeepFace
import os
import shutil
import random

_all__ = ['process', 'inference']