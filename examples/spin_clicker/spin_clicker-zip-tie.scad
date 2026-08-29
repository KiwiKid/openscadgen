

	include <BOSL2/std.scad>;
include <BOSL2/joiners.scad>;
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

partType = "all";

spinnerWidth = 2;
spinnerLength = 80;
spinnerThickness = 30;
spinnerRounding = 1;

holderWidth = 18;

holderWall = 12;
holderHeight = 40 ;

snakePieceThickness = 25;

snakeCurveLength = 0; 

spinnerOffset = 18;


doveTailHeight = 22;
doveTailDepth = 23;
doveTailWidth = 8;

doveTailOffset = 2;
wedgeGap = 12;
attachmentPointSize = [11,doveTailDepth,24];

attachmentPointMove = [0,2,-12.5];

doveTailMove = [0,18,-22];
plugScale = [0.99,1,0.99];

joinMove = [0,0.48,-10];

tipMove = [0,2,-14];

tipLength = 15;


module screw_holder(){
 rotate([90,0,0]){
            difference(){
                cyl(d=holderWidth, h=spinnerThickness+holderHeight, rounding1=3);
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
	module spin_clicker(onlyOne=false){



		//#cuboid([spinnerWidth,spinnerLength,spinnerThickness])
        difference(){
           union(){


            if(!onlyOne){
            up(5)
            fwd(holderHeight/2)
            screw_holder();
            
            }
            up(holderWidth)
            fwd(holderHeight/2)
            screw_holder();
            
            
               difference(){
               union(){
               move(joinMove)
                move(attachmentPointMove)
                dovetail("male", width=attachmentPointSize[0], height=attachmentPointSize[1], slide=attachmentPointSize[2]);
//                #cuboid(attachmentPointSize, rounding=2);
               
               move(joinMove+tipMove)
               
                    cuboid([3,doveTailDepth,tipLength], rounding=1);
                    
                    }
                    
               move(joinMove+tipMove)
                   back(2)
                    cuboid([1.1,doveTailDepth,tipLength+10]);
                }
                
           };
           
           move(joinMove)
           down(doveTailHeight/2)
          back(5)
          down(1)
            dovetail("male", width=doveTailWidth, height=doveTailHeight, slide=doveTailDepth);
        }
           
	}


    sliced(renderType=renderType) {
        if(partType == "frame" || partType == "all"){

            spin_clicker();
        }
        /*
        if (partType == "plug" || partType == "all"){

            difference() {
                
            
           down(doveTailHeight/2)
            dovetail("male", width=doveTailWidth, height=doveTailHeight, slide=doveTailDepth);

        up(doveTailHeight/2)
        rotate([90,0,0]){
        
        up(holderWidth+1.8)
               cyl(d=holderWidth-holderWall, h=spinnerThickness+holderHeight+0.001, chamfer1=-2);
        }
            }
        }*/

        if(partType == "plug" || partType == "all"){
          //  down(doveTailHeight+5)
         //   intersection() {
                translate(doveTailMove)
                scale(plugScale)
                 dovetail("male", width=doveTailWidth, height=doveTailHeight, slide=doveTailDepth-wedgeGap);

           //     #spin_clicker(onlyOne=true);
            //}
        }
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

