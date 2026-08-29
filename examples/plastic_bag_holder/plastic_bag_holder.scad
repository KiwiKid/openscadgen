

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
	renderType = "obj";


    partType = "all"; // "lid"
    
    
    globalScale = 0.8;
    partSeperation = 40;

    diameter = 120;
    innerDiameter = 110;
    screwType = "#12";
    
    baseHeight = 20;
    
    scaleUp = 4;
    
    baseBottomDown = 15;
    
    
    
    
    clipWidth = 8;
    clipHeight =18;
    clipDepth = 2;
    
    wedgeLength = 4;
    
    wedgeBump = 0.5;
    wedgeBumpDown = -3;
    
    clipOffset = 1.15;
    clipDepress = 6;
    
    screwPlateWidth = 30;
    screwPlateDepth = 5;
    screwPlateHeight = 30;
    
    module clip(){
            
            cuboid([clipDepth,clipWidth,clipHeight], rounding=0.5)
            
            down(wedgeBumpDown)
            fwd(clipWidth/2)
             zrot(-90)
            attach(BOT)
            right(clipWidth/2)
            back(wedgeBump)
//           hull() {
///                wedge([clipWidth, wedgeBump, wedgeLength]);
                xscale(1)
                sphere(r=2);
  //              }
            
           /* zrot(-90)
            down(clipDepth*1.08)
            back(clipDepth/2)
            wedge([clipWidth, wedgeBump, wedgeLength], anchor=CENTER);*/
            }


	module plastic_bag_holder(){

        if (partType == "lid" || partType == "all") {
            difference() {
		    
            
                cyl(h=15, d=innerDiameter, chamfer2=-7, chamfer1=3);
                
                        
                    
             //   #screw_hole(screwType, head="flat", l=5, thread=true);

              //  up(0.0001)
             //   scale([0.65,0.65,1.1])
             //   screw_hole(screwType, head="flat", thread=false, l=5);
             
                cyl(h=30, d=innerDiameter*0.9, chamfer2=-15);
            }
            
            for (i = [30 : 90 : 330]){
            
                zrot(i)
                down(clipDepress)
                left(innerDiameter/2-clipOffset)
                zrot(180)
                #clip();
            }
            

        }
        
                

        down(partSeperation)
        if (partType == "base" || partType == "all") {
           /* difference() {
               cyl(h=20, d=diameter);
                
                
                cyl(h=101, d=diameter*0.7);
            }*/
            
         //   down(baseBottomDown)
            difference(){

               union(){
                cyl(h=15, d=innerDiameter*1.1);
               
               move([-screwPlateWidth/2,-innerDiameter/2-screwPlateDepth/2-0.4,-screwPlateHeight/1.3])
                cuboid([screwPlateWidth,screwPlateDepth,screwPlateHeight], anchor=BOTTOM+LEFT, rounding=5, edges="Y");
                }

               
               cyl(h=16, d=innerDiameter+1, chamfer2=-3);
               

               
               // hole for screw
         //       up(30)
           //    back(diameter+1)
           fwd(112)
           down(screwPlateHeight/2)
               rotate([90,0,0])
           //    down(20)
               union(){
                   right(10)
                   cyl(h=120, d=4, chamfer1=-6);
                   
                   left(10)
                   cyl(h=120, d=4, chamfer1=-6);
               }
               
  
                
                rotate([90,0,0])
               cyl(h=40, d=12, anchor=BOTTOM);
               }
               
            //   down(baseBottomDown)
        }
	}


    sliced(renderType=renderType) {
    scale(globalScale)
        plastic_bag_holder();
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

