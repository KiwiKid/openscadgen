

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
    
    
    mainSize = 135;
    cutoutMove  = [10,10,1];
    
    bottomPlateHeight = 2;
    plateSize = 70;
    holderHeight = 25;
    
    cutoutHeightOffset = 2;
    cuoutXYMoveOffset = -67;
    radusOffset = 15;
    
    holderRadius = 100;

	module circular_holder(){

        sphereRadius = 6;
    difference(){
        
    cuboid([mainSize,mainSize,holderHeight], rounding=10, anchor=BOTTOM, edges = "Z");
    
    move([holderRadius+cuoutXYMoveOffset,-holderRadius-cuoutXYMoveOffset,bottomPlateHeight+0.01])
    #cyl(r=holderRadius, height=holderHeight-cutoutHeightOffset, anchor=BOTTOM, rounding=6);
    
    move([46,-plateSize/2,bottomPlateHeight])
    rotate([0,0,45])
    cuboid([200,101,20], anchor=BOTTOM);
}
		

	}
    

    sliced(renderType=renderType) {
        scale([0.5,0.5, 1])
        circular_holder();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 7],
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

