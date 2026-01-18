

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

    holeCount = 5;
    holeSize = 6;
    cylChamfer = -3;
    holeOffset = 10;
    holeDownOffset = 10;
    gapBetween = 20;
    
    cuboidLength = 40;//200;
    cuboidDepth = 40;
    cuboidHeight = 4;
    
    
    module screwHole(){
    cyl(r=holeSize/2, h=cuboidHeight+0.001, chamfer2=cylChamfer);
    }

	module screw_in_wedge(){
        difference() {
            cuboid([cuboidLength,cuboidDepth,cuboidHeight], rounding=2, anchor=LEFT);

           
            fwd(holeDownOffset)
           // xdistribute(spacing=gapBetween){
                for (i = [0:holeCount-1]) {
                      right(gapBetween/2+holeSize/2)
                      xmove((i * gapBetween)) {
                        #screwHole();
                      }
                    }
           // }

        }
		
	}


    sliced(renderType=renderType) {
        screw_in_wedge();
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

