from bs4 import BeautifulSoup
import urllib2

def foo():
    page = urllib2.urlopen("https://www.google.com/").read()
    soup = BeautifulSoup(page, 'html.parser')
    return soup.title.string
