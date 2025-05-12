

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;


	module furniture_foot(){

        n = 64;            // number of points per circle
        r1 = 10;           // radius of bottom circle
        r2 = 5;            // radius of top circle
        h = 20;            // height between the circles

        // Convert 2D circles to point lists
        bottom_circle = circlepoints(r1, n);                  // at z = 0
        top_circle    = translate_points(circlepoints(r2, n), [0, 0, h]);  // at z = h

        // Skin the circles into a 3D shape
        skin([bottom_circle, top_circle], closed=true);  // closed=true caps both ends

	}


    sliced(renderType="") {
        furniture_foot();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.3,
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

