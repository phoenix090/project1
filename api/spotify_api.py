import spotipy, sys, os
from spotipy.oauth2 import SpotifyClientCredentials
import spotipy.util as util

scope = 'user-library-read'

client_credentials_manager = SpotifyClientCredentials()
username = os.getenv('USERNAME')
token = util.prompt_for_user_token(username, scope)
if not token:
    print('Something went wrong, exiting')
    exit(1)

sp = spotipy.Spotify(token, True, client_credentials_manager=client_credentials_manager)
user = sp.user(username)

if not sp:
    print('Something went wrong, exiting')
    exit(1)


# API call to get the current user info
# GET: 200 
def get_user(username):
    user = sp.user(username)
    return user

# GET: retrieve an artist by ID
''' TODO: make the id dynamically '''
def get_artist(choice):
    valid=['album','single','appears_on','compilation']
    song_list= []
    if choice in valid:
        artist = sp.artist_albums('3TVXtAsR1Inumwj472S9r4',album_type=choice,limit=20)
        for item in artist['items']:
            song={}
            song['name'] = item['name']
            song['release_date'] = item['release_date']
            song_list.append(song)
    else:
        artist = sp.artist_albums('3TVXtAsR1Inumwj472S9r4',album_type='album',limit=20)
        for item in artist['items']:
            song={}
            song['name'] = item['name']
            song['release_date'] = item['release_date']
            song_list.append(song)
    return song_list

# GET: gets an artist's top tracks, use COUNTY to limit the request
def get_artist_top_tracks():
    top_tracks = sp.artist_top_tracks('2YZyLoL8N0Wb9xBt1NhZWg')
    for tracks in top_tracks['tracks']:
        for album in tracks:
            if album == 'album':
                #print('name: ', tracks['album']['name'], '\n')
                #print('total tracks: ', tracks['album']['total_tracks'], '\n')
                #print('release date: ', tracks['album']['release_date'], '\n')
                print('key: ', tracks['album'], '\n')   

# Gets detailed information about the current user
def get_current_user_detail():
    var=sp.me()
    return(var['display_name'])
    #return sp.me()

# gets the current users top tracks
def get_current_user_top_tracks():
    tracks = sp.current_user_top_tracks(limit=5)
    print(tracks)
    return tracks

# Get information about the current users currently playing track
def user_playlist(user=None, cap=5):
    if user == 'none':
        user = username
    playlists = sp.user_playlists(user, limit=int(cap))
    plists = {}
    plists['total'] = playlists['total']
    plists['limit'] = playlists['limit']
    p = {}
    ant = 1
    #print(playlists['items'])
    items = playlists['items']
    for item in items:
        #print(item, '\n\n')
        p['name'] = item['name']
        p['owner'] = item['owner']['display_name']
        p['id'] = item['id']
        p['total_tracks'] = item['tracks']['total']
        p['uri'] = item['uri']
        key = 'playlist_' + str(ant)
        plists[key] = p 
        ant += 1
    return plists
    

