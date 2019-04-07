$(document).ready(function() {

    $.ajax("/api/queue", function(data) {
        data.each(function(vid) {
            $("#vids").append($("<div>").append($("<span>").text(vid.url)))
        })
    })

})