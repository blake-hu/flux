from setuptools import setup, find_packages

setup(
    name='inference',  # Replace 'your_module' with the name of your module
    version='1.0.1',  # Specify the version of your module
    author='Your Name',  # Replace 'Your Name' with your name or the author name
    author_email='your@email.com',  # Replace 'your@email.com' with your email
    description='A short description of your module',  # Add a short description
    long_description='A longer description of your module',  # Add a longer description if needed
    long_description_content_type='text/markdown',  # Specify the type of the long description
    url='https://github.com/your_username/your_module',  # Add the URL of your module's repository
    packages=find_packages(),  # Automatically find all packages and subpackages
    classifiers=[  # Add classifiers to specify the audience and maturity of your module
        'Development Status :: 3 - Alpha',
        'Intended Audience :: Developers',
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.9',
    ],
    python_requires='>=3.9',  # Specify the Python versions supported by your module
    install_requires=["numpy", "pandas", "opencv-python", "scikit-learn", "deepface"]  # Add any dependencies required by your module
)
