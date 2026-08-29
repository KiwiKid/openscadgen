
// file::///./config.toml
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
innerRadius = 4;

cylHeight = 20;
wallSize = -1;// 1;

outerRadius = innerRadius + wallSize;

cupInnerRoundiung = 2.5;
cupOuterRounding1 = 0;
cupOuterChamfer1 = -6.5;
starStepCount = 3;

floorDepth = 1;
subMode = "sizer";


mode =  "dowel";// "decagon";// "dowelPattern";

module decagon(){
    difference(){
            cyl(r=outerRadius, h=cylHeight);
            
                up(wallSize)
                linear_extrude(h=cylHeight)
            regular_ngon(n=6, ir=innerRadius);
          }
}

	module simple_cylinder(mode = "tube"){
    if(mode == "tube"){
        difference(){
            cyl(r=outerRadius, h=cylHeight);
            
           cyl(r=innerRadius, h=cylHeight+0.1);
           }
    } else if(mode == "decagon"){
    
    if(subMode == "sizer"){
    intersection(){
        decagon();
        up(5)
        cuboid([100,100,1]);
        }
     } else {
        decagon();
     }
       
    }else if(mode =="cup") {
            difference(){
                cyl(r=outerRadius, h=cylHeight);
                
                up(wallSize)
                cyl(r=innerRadius, h=cylHeight+0.1, rounding=cupInnerRoundiung);
           }    
        
	} else  if(mode =="dowelPattern") {
            difference(){
               
                cyl(r=outerRadius, h=cylHeight, chamfer1=cupOuterChamfer1, chamfer2=-1);
                
 //               / Inner cutout with vertical dowel-like ribs lifted by 10 units
            up(wallSize+floorDepth) {
                linear_sweep(
                    star(r=innerRadius, n=24, step=starStepCount), 
                    h=cylHeight+0.1, 
                    center=true
                );
                down(wallSize+1)
                cyl(r=innerRadius*0.9, h=cylHeight-wallSize-floorDepth+0.001, rounding=cupInnerRoundiung, chamfer2=-1);
                }
           }    
        
	} else if(mode =="dowel") {

                linear_sweep(
                    star(r=innerRadius, n=24, step=starStepCount), 
                    h=cylHeight+0.1, 
                    center=true
                );
                down(wallSize+1)
                #cyl(r=innerRadius*0.96, h=cylHeight*1.1, rounding=cupInnerRoundiung);
                
           
        
	} else {
               assert(false, "should have a valid mode");         
        }
        }
    
    
    simple_cylinder(mode=mode);

