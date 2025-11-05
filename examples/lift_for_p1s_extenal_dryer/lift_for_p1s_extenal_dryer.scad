

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
    tubeLength = 240;
    rectLength = 240;
    rectWidth = 100;
    rectSize = [rectWidth,rectLength];
    wallSize= 5;
    
    bottomFootLength = 30;
    bottomFootHeight = 32;
    bottomFootMove = [-40,rectLength/2-bottomFootHeight/2+4,-3];
    
    topCutoutMove = [0, -rectLength/2-6, tubeLength/2];
    topCutoutSize = [rectWidth*0.9, 10, tubeLength*0.9];
	module lift_for_p1s_extenal_dryer(){
            tex = texture("trunc_ribs");

    difference(){
        linear_sweep(
            rect(rectSize), h=tubeLength, texture=tex,
            tex_depth=3, tex_size=[10,10],
            style="concave"
        );
        
//            down(wallSize)
down(1)
fwd(bottomFootHeight/2)
            linear_sweep(
                rect([rectWidth-wallSize, rectLength-bottomFootHeight-wallSize]), h=tubeLength*1.1
            );
            
            move(topCutoutMove)
            cuboid(topCutoutSize);
            
             
            move(bottomFootMove)
             linear_sweep(
                rect([bottomFootLength, bottomFootHeight]), h=tubeLength*1.1, texture=tex,
                    tex_depth=3, tex_size=[10,10],
                    style="concave"
            );
            
        }
           
        
	}


    sliced(renderType=renderType) {
        lift_for_p1s_extenal_dryer();
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

