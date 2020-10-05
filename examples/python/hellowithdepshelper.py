from bs4 import BeautifulSoup
import urllib.request

def foo(event, context):
    page = urllib.request.urlopen("https://www.google.com/").read()
    soup = BeautifulSoup(page, 'html.parser')
    return soup.title.string
