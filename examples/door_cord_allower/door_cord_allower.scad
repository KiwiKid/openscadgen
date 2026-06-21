

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
    
    cordHoleRadius = 3.5;
    
    
    cordHoleOffset = 1.2;
        width  = 130;
    length = 60;
    height = 6.8;
    cordHoleApart = 10;

    module cord_hole(){
        up(cordHoleOffset)
        rotate([90,0,90])
        cyl(r=cordHoleRadius,h=width*1.2);
        }

	module door_cord_allower(){


            difference(){
            intersection() {
                scale([width/2, length/2, height])
                    sphere(r=1, $fn=100);

                translate([-width/2, -length/2, 0])
                    cube([width, length, height]);
                    
            }
            cord_hole();
            
            fwd(cordHoleApart)
            cord_hole();
            
          back(cordHoleApart)
            cord_hole();
            }
	}


    sliced(renderType=renderType) {
        door_cord_allower();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
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

