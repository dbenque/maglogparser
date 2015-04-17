package magLogParserServer

import "text/template"

//GetPageTemplate return the templaet with the page
func GetPageTemplate() (tmpl *template.Template, err error) {

	tmpl, err = template.New("page").Parse(`
{{define "PAGE"}}
<html>
<head>
<script src="http://ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js"></script>
<script src="http://code.highcharts.com/highcharts.js"></script>
</head>
<body>
<script src="http://code.highcharts.com/highcharts.js"></script>
<script src="http://code.highcharts.com/modules/exporting.js"></script>

<div id="container" style="min-width: 310px; height: 400px; margin: 0 auto"></div>

<script>
$(function () {

  var chart;

  function requestData() {
    $.ajax({
        url: 'http://{{.}}/data',
        success: function(points) {

            chart.series[0].setData(points["Max"],true,true,false)
            chart.series[1].setData(points["Min"],true,true,false)
            chart.series[2].setData(points["Average"],true,true,false)

            // call it again after one second
            setTimeout(requestData, 1000);
        },
        error : function(resultat, statut, erreur){
          console.log(resultat)
          console.log(status)
          console.log(erreur)
          window.alert("error")
       },
        cache: false
    });
}


    $(document).ready(function () {
        Highcharts.setOptions({
            global: {
                useUTC: false
            }
        });

        chart = new Highcharts.Chart({
            chart: {
                renderTo: 'container',
                type: 'spline',
                animation: Highcharts.svg, // don't animate in old IE
                marginRight: 10,
                events: {
                    load: requestData
                }
            },
            title: {
                text: 'Live random data'
            },
            xAxis: {
                type: 'datetime',
                tickPixelInterval: 150
            },
            yAxis: {
                title: {
                    text: 'Value'
                },
                plotLines: [{
                    value: 0,
                    width: 1,
                    color: '#808080'
                }]
            },
            legend: {
                enabled: true
            },
            exporting: {
                enabled: true
            },
            series: [{
                name: 'Max',
                data: []
            },{
                name: 'Min',
                data: []
            },{
                name: 'Median',
                data: []
            }]
        });
    });
});

</script>
</body>
</html>
{{end}}
`)
	return
}
