

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

    markerRotate = 45;
    markerMove = [0,0,0];
    markerThick = 6;
    markerSize = [100, 10, markerThick];
    
    hookWallSize = 20;
    hookInset = -7;
    hookWidth = 60;
    hookThick = 40;
    hookSize = [hookWidth, 50, hookThick];
    
    hookInnerSize = [200, 15, 30];
    
    module hook(){
    difference(){
        cuboid(hookSize, rounding=5);
        
        back(hookInset)
        cuboid(hookInnerSize, rounding=1);
        
}
    }

      module xMarker(){
            zrot(markerRotate)
            move(markerMove)
            cuboid(markerSize);
                        
            zrot(-markerRotate)
            move(markerMove)
            cuboid(markerSize);
      }
    
	module dont_use_door_indicator(){

		hook();
        down(hookThick/2+markerThick/2)
        xMarker();
        
	}


    sliced(renderType=renderType) {
        dont_use_door_indicator();
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
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
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

