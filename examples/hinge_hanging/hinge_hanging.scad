

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
    
    topExtend = 10;
    
// The heart path flattened out completely (Z coordinates dropped or set to 0)
flat_heart_path = [
    [10,   45+topExtend],  // 1. START: Center top dipped cusp
    [20,  55+topExtend],  // 2. Right top lobe
    [30,  50+topExtend],  // 3. Right wide shoulder
    [30,  20],  // 3. Right wide shoulder
    [20,  5],  // 4. Right lower curve
    [12,    0],  // 5. Bottom flat base (Right side)
    [-12,   0],  // 6. Bottom flat base (Left side)
    [-20, 5],  // 7. Left lower curve
    [-30, 20],  // 7. Left lower curve
    [-30, 50+topExtend],  // 8. Left wide shoulder
    [-20, 55+topExtend],   // 9. END: Left top lobe
    [-10, 45+topExtend]   // 9. END: Left top lobe
];

// Flat scaling factors to keep your thickened base logic intact
thickness_multipliers = [
    1.0,  // 1. Cusp (Normal)
    1.0,  // 2. Lobe
    1.2,  // 3. Shoulder
    1.8,  // 4. Lower Curve
    2.5,  // 5. BASE RIGHT (Thickened)
    2.5,  // 6. BASE LEFT (Thickened)
    1.8,  // 7. Lower Curve
    1.2,  // 8. Shoulder
    1.0,   // 9. Lobe
    1.0
]; 

	module hinge_hanging(){

 path_sweep(
        shape=rect([5, 6], rounding=1, center=true), 
        path=flat_heart_path, 
      //  scale=thickness_multipliers, 
        closed=false, 
        caps=true
    );
    }


    sliced(renderType=renderType) {
        hinge_hanging();
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

