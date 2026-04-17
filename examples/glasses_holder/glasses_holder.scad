

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

    
   mainRotate = 35;
   
   cutoutWidth = 8;
   
   globalScale = 0.5;
   
   mainDiameter = 30;

   mainHeight = 100;
   
	module glasses_holder(){
		intersection() {
			// Non-rotated square boundary
            up(50)
			cuboid([250, 250, 180]);

			// Rotated cylinder with groove on top
			xrot(mainRotate) {
				difference() {
					// Main cylinder
					cyl(d=mainDiameter, h=mainHeight,rounding=mainDiameter/3);

					// Groove cutout sits on top, un-rotated relative to world
                    back(3)
					up(mainHeight/2-8)
						xrot(-mainRotate) // cancel parent rotation
                            yrot(90)
                            union(){
                                
                               xcopies(4, n=cutoutWidth)
                                cyl(d=cutoutWidth, h=100);

                            }
				}
                
            
            yscale(2)
            move([0, -mainHeight/10, -35])
            
            sphere(mainDiameter/2);
			}
		}
	}


    sliced(renderType=renderType) {
    scale(globalScale)
        glasses_holder();
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

