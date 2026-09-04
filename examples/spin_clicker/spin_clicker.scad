

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
spinnerLength = 90;
spinnerThickness = 3;
spinnerRounding = 1;

holderWidth = 18;
holderOuterWidth = 10;
holderOuterGapOffset = 1;
holderClickerBarOffset=  2;

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

 // a series of points that will create a snaking curve back and forth
        module snakePiece(height=spinnerThickness){
                path = [for (a=[0:30:900]) [a-180, 60*sin(a)]];

        down(spinnerThickness/2)
            scale([0.15,0.04,1])
            rotate([0,0,90])
                linear_extrude(height=spinnerThickness){
                stroke(path, width=snakePieceThickness, endcaps="round", joint_angle=0, joint_width=5);
            };
        }
	module spin_clicker(){


       // fwd(spinnerLength/2+snakeCurveLength+spinnerOffset)
      // snakePiece();

		cuboid([spinnerWidth,spinnerLength,spinnerThickness], rounding=.3)

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

