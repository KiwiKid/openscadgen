

	include <BOSL2/std.scad>;

include <BOSL2/screws.scad>;
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
	renderType = "vertSlice";


	module le_bottle_lamp(){
		
        holeXYScale = 4.3;
        holeScale = 4.8;
        baseDepth = 30;
        baseWidth = 100;
        baseRounding=30;
        cordHole = baseWidth*1.1;
        cordDiameter = 6;
        cordDown = 12;
        cordScale = [1.5,1.5,1];
        
        screwXYScale = 1.3;
        screwZScale = 1.6;
    difference(){
    
      cyl(baseDepth, baseWidth, rounding2=baseRounding);
      
     
                        centerHoleSpec = [["system","ISO"],
        ["type","screw_info"],
        ["pitch", 1.3],
        ["head", "none"],
        ["head_size", 1],
        ["head_size_sharp", 22],
        ["head_angle", 60],
        ["diameter",9.8]];
        
        up(baseDepth/2)
        scale([holeXYScale,holeXYScale,holeScale])
        screw_hole(centerHoleSpec, l=4, anchor=TOP,thread=true,spin=180);
  
        scale(cordScale)
        down(cordDown)
        fwd(baseWidth/2+8)
        rotate([0,80,90])
        #cyl(cordHole, cordDiameter);
      }
      
      difference(){
                spec = [["system","ISO"],
        ["type","screw_info"],
        // More turns over the same height => smaller pitch (turns = l/pitch).
        // Keep `l` the same; just change `screwTurns` (or set pitch directly).

        ["head", "flat"],
        ["flat_height",1],
        ["head_size", 20],
        ["head_size_sharp", 22],
        ["head_angle", 60],
        ["diameter",21]];

        screwLen = 150;
        screwTurns = 40;          // increase for more turns, decrease for fewersc
        
        
        down(baseDepth+10)
        // NOTE: In BOSL2, the ",<...>" suffix in the spec string is a length (in inches), not pitch.
        // For a long "wrap path" (eg lights), override pitch directly with `thread=<pitch_mm>`.
        // `starts` increases lead (advance per revolution) without changing pitch height.
        
        scale([screwXYScale,screwXYScale,screwZScale])
        screw(spec,thread=7, l=screwLen, orient=DOWN, atype="head", anchor=TOP);
        
        down(baseDepth/2)
        cuboid(100, anchor=TOP);
        }
       }


    sliced(renderType=renderType) {
        le_bottle_lamp();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    horzSliceRotate = 90,
    vertSlicePos = [0, -500, -500],
    vertSliceRotate = 30
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
            rotate([0,0,vertSliceRotate])
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                rotate([0,0,vertSliceRotate])
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

