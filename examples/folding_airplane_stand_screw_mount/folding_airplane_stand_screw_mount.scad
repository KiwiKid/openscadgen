

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

    mountLength = 100;
    
    mountScrewOffset = 30;
    mountScrewRadius = 8;
    
    mountHeight = 30;
    mountDepth = 21;
    screwRecess = 10;
    mountRounding = 7;
    title = " Choose Happy";
    
    module screwHole(){
        down(screwRecess)
        cyl(h=mountHeight+0.1,r=3,chamfer2=-2);

        up(mountHeight-screwRecess)
        cyl(h=mountHeight+0.1,r=6);
    }
    
	module folding_airplane_stand_screw_mount(){
    difference(){
		cuboid([mountDepth,mountLength,mountHeight], rounding=mountRounding, edges=TOP);
        
            fwd(mountScrewOffset)
            screwHole();
            
            back(mountScrewOffset)
            screwHole();
            
            rotate([90,0,90])
            translate([-mountLength/2,-10,10])
            text3d(title);
        
        }
	}


    sliced(renderType=renderType) {
        folding_airplane_stand_screw_mount();
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

