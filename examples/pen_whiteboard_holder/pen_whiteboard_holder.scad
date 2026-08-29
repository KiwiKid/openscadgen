

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

  magnetHoleDiameter = 6.05;
  magnetHoleDepth = 2.5;
  magnetHoleOffset = 10;

 // penDiameter =9.98;
  //penHolderDown = 0.05;
  penDiameter = 11.5;
  penHolderDown = -0.7;
  
  penHolderWallSize = 3;

  holderHeight = 10;
holderWidth =30;
holderDepth = 40;

  magnetHoleZOffset = holderHeight/2-(magnetHoleDepth/2)+0.001;


	module pen_whiteboard_holder(){
		difference(){
            union(){
                cuboid([30,10,10], rounding=3);
                
                down(penHolderDown)
                rotate([90,0,0])
                cyl(h=holderHeight*0.9-0.01, d=penDiameter+penHolderWallSize, center=true, rounding=1);
            }
            down(10)
            #cuboid([30,10,10]);
            
            down(penHolderDown)
            rotate([90,0,0])
            cyl(h=holderHeight+0.01, d=penDiameter, anchor=CENTER);

            left(magnetHoleOffset)
            down(magnetHoleZOffset)
            cyl(h=magnetHoleDepth, d=magnetHoleDiameter, anchor=CENTER);

            right(magnetHoleOffset)
            down(magnetHoleZOffset)
            #cyl(h=magnetHoleDepth, d=magnetHoleDiameter, anchor=CENTER);
        }
        
	}


    sliced(renderType=renderType) {
        pen_whiteboard_holder();
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

