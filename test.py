import pyttsx3
import argparse
parser = argparse.ArgumentParser()
parser.add_argument('--text')
args = parser.parse_args()
engine = pyttsx3.init()
engine.say(args.text)
engine.runAndWait()
exit
