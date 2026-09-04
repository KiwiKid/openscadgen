

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

spinnerWidth = 1.5;
spinnerLength = 92;
spinnerThickness = 3;
spinnerRounding = 1;
flatBottomHeight = -2.5;

holderWidth = 18;
holderOuterWidth = 13;
holderOuterGapOffset = 1;
holderClickerBarOffset=  2.1;

holderWall = 12;
holderHeight = 40 ;

snakePieceThickness = 25;

snakeCurveLength = 0; 

spinnerOffset = 18;


module screw_holder(){
 rotate([90,0,0]){
            difference(){
                cyl(d=holderOuterWidth, h=spinnerThickness+holderHeight, rounding1=3);
                cyl(d=holderWidth-holderWall, h=spinnerThickness+holderHeight+0.001, chamfer1=-2);
                
            }
            
            }

}

    textSize = 5;
    textDepth = 6; // Extrusion depth cutting into the wall
  //  cylRadius = 43; // Distance from cylinder center to text surface
    textRadius = 8.7;
        gap = holderOuterWidth-1; // Distance between cylinder centers

textCutoutLengthOffset = 1.8;

module textGroup(textRotate=90, skipIndex=3) {

    
    
    zrot_copies(n=4, r=textRadius, subrot=true) {
    if($idx != skipIndex){
    rotate([0,textRotate,0])
        text3d("Spin to Win", size=textSize, h=textDepth, font="Arial Rounded MT Bold", center=true);
        }
    }
}
module textCutout() {
    
    textCenterRotate = [0, 0, 0];
    textCenterMove = [0, -spinnerLength/2 - holderOuterWidth+textCutoutLengthOffset, -holderHeight/2];

    rotate(textCenterRotate)
    move(textCenterMove) {
        // Top cylinder group
        translate([0, gap / 2, 0])
        textGroup(skipIndex=3);
        
        // Bottom cylinder group (Mirrored)
        translate([0, -gap / 2, 0])
        mirror([0, 0, 0])
        textGroup(skipIndex=1);
    }
}
	module spin_clicker(){

    difference(){
    intersection(){
       // fwd(spinnerLength/2+snakeCurveLength+spinnerOffset)
      // snakePiece();

		cuboid([spinnerWidth,spinnerLength,spinnerThickness], rounding=.3, anchor=BOT)

        up(spinnerThickness/3.5)
            fwd(holderOuterWidth-holderClickerBarOffset)
            attach(FWD){         
                up(holderOuterGapOffset)
            up(holderOuterWidth/2-holderOuterGapOffset)
            fwd(holderHeight/2)
          //  up(snakeCurveLength*2+spinnerOffset)
            screw_holder();
            
            
            up(-holderOuterWidth/2+holderOuterGapOffset)
            fwd(holderHeight/2)
       //     up(snakeCurveLength*2+spinnerOffset)
            screw_holder();
            }
          
          flatBottomBoxSize = [1000,1000,100];
          
          down(flatBottomHeight)
          cuboid(flatBottomBoxSize, anchor=TOP);
          }
          textCutout();
	}
    
    }


    sliced(renderType=renderType) {
        spin_clicker();
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

