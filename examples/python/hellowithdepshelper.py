from bs4 import BeautifulSoup
import urllib3

def foo(event, context):
    page = urllib3.urlopen("GET", "https://www.google.com/").read()
    soup = BeautifulSoup(page, 'html.parser')
    return soup.title.string
