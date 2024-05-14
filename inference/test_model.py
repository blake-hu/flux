import pickle
from sklearn.linear_model import LinearRegression
from sklearn.datasets import make_regression

# Generate some sample data for demonstration
X, y = make_regression(n_samples=100, n_features=1, noise=0.1)

# Train your model
model = LinearRegression()
model.fit(X, y)

# Save the trained model to a file
with open('linear_regression_model.pkl', 'wb') as file:
    pickle.dump(model, file)
