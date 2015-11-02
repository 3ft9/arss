(function updateTable() {
	$.ajax({
		url: '/?ajax=1',
		success: function(data) {
			$('#feedstable').html(data);
		},
		complete: function() {
			setTimeout(updateTable, 5000);
		}
	});
})();
