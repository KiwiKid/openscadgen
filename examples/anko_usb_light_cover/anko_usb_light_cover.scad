

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";

    wallDepth=0.5;
radiusBottom=11.95;
radiusTop=11.95;  //1.8 -> 2.1
height=24;
topRounding=4;
bottomHoldingRounding=1;
letter="W";
textSize=9;

	module anko_usb_light_cover(radiusTop=radiusTop, radiusBottom=radiusBottom,wallDepth=wallDepth, height=height, topRounding=topRounding, letter=letter, bottomHoldingRounding=bottomHoldingRounding){
		difference(){
            #cyl(r2=radiusTop+wallDepth, r1=radiusBottom+wallDepth, h=height, rounding2=topRounding);
            
            down(wallDepth)
            cyl(r2=radiusTop, r1=radiusBottom, h=height, rounding1=bottomHoldingRounding, rounding2=topRounding);
            
            if(letter != ""){
            up(height/2-0.002)
                #text3d(text=letter, center=true, size=textSize, height=10);
            }
        }
        
	}


    sliced(renderType=renderType) {
        anko_usb_light_cover();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.5,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

