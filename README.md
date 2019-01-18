# Radiotation

Make a radio, rotate it among friends. Uses the Spotify API to load a list of songs for a given search query. Each person has a queue (based on their cookie-defined session), and repeatedly calling the `/pop` endpoint will one song from the current queue, and rotate through all the available non-empty queues.

# TODO
- Better logging
