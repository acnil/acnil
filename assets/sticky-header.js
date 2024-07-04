function getAllLevel2Headers () {
  // Select all section elements
  var sections = document.getElementsByTagName('section');

  // Initialize an array to hold the result
  var headersArray = [];

  // Loop through each section
  for (var i = 0; i < sections.length; i++) {
    // Get all h2 elements within the current section
    var headers = sections[i].getElementsByTagName('h2');

    // Loop through each header and push an object with text content and offsetTop
    for (var j = 0; j < headers.length; j++) {
      headersArray.push({
        text: headers[j].textContent,
        offsetTop: headers[j].offsetTop
      });
    }
  }

  return headersArray;
}

function getClosestHeader () {
  var headers = getAllLevel2Headers();
  var scrollPosition = window.scrollY
  var closestHeader = null;

  for (var i = 0; i < headers.length; i++) {
    if (headers[i].offsetTop < scrollPosition) {
      closestHeader = headers[i]
    } else {
      break
    }
  }

  return closestHeader;
}

function watchScroll () {
  var header = getClosestHeader()
  if (header) {
    document.getElementById("current-section").innerText = header.text
  } else {
    document.getElementById("current-section").innerText = ""
  }
}
window.onscroll = watchScroll