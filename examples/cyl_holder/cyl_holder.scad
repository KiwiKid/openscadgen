

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


	module cyl_holder(holderDiameter=7, holderInnerDiameter=6, holderHeight=20, screwDiameter=5){
		difference(){
            
            union(){
                cyl(h=holderHeight, d=holderDiameter);
                cuboid([holderDiameter*3,holderDiameter,holderDiameter], rounding=0.5);
            }
            
            cyl(h=holderHeight+10, d=holderInnerDiameter);
            

            rotate([0,90,90])
            back(holderDiameter)
            #cyl(h=holderDiameter, d=screwDiameter, chamfer2=-1);
            
            
            rotate([0,90,90])
            fwd(holderDiameter)
            #cyl(h=holderDiameter, d=screwDiameter, chamfer2=-1);

        }
	}


    sliced(renderType=renderType) {
        cyl_holder();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.1,
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

